package search

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Index struct {
	db *sql.DB
}

type SearchResult struct {
	Path    string
	Title   string
	Snippet string
	Rank    float64
}

func Open(dbPath string) (*Index, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %w", err)
	}

	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open index: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Index{db: db}, nil
}

func createSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY,
			path TEXT UNIQUE NOT NULL,
			title TEXT,
			mod_time INTEGER NOT NULL,
			indexed_at INTEGER NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_notes_path ON notes(path);
		CREATE INDEX IF NOT EXISTS idx_notes_mod_time ON notes(mod_time);

		CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(
			title,
			content,
			content='',
			tokenize="trigram"
		);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func (idx *Index) Close() error {
	if idx.db != nil {
		return idx.db.Close()
	}
	return nil
}

func (idx *Index) IndexNote(ctx context.Context, path, title, content string, modTime time.Time) error {
	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existingID int64
	var existingModTime int64
	err = tx.QueryRowContext(ctx,
		"SELECT id, mod_time FROM notes WHERE path = ?",
		path,
	).Scan(&existingID, &existingModTime)

	newModTime := modTime.Unix()
	now := time.Now().Unix()

	if err == sql.ErrNoRows {
		result, err := tx.ExecContext(ctx,
			`INSERT INTO notes (path, title, mod_time, indexed_at) VALUES (?, ?, ?, ?)`,
			path, title, newModTime, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert note: %w", err)
		}

		noteID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get note ID: %w", err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO notes_fts(rowid, title, content) VALUES (?, ?, ?)`,
			noteID, title, content,
		); err != nil {
			return fmt.Errorf("failed to insert into FTS: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query existing note: %w", err)
	} else if existingModTime != newModTime {
		if _, err := tx.ExecContext(ctx,
			`UPDATE notes SET title = ?, mod_time = ?, indexed_at = ? WHERE id = ?`,
			title, newModTime, now, existingID,
		); err != nil {
			return fmt.Errorf("failed to update note: %w", err)
		}

		if _, err := tx.ExecContext(ctx,
			`DELETE FROM notes_fts WHERE rowid = ?`,
			existingID,
		); err != nil {
			return fmt.Errorf("failed to delete from FTS: %w", err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO notes_fts(rowid, title, content) VALUES (?, ?, ?)`,
			existingID, title, content,
		); err != nil {
			return fmt.Errorf("failed to insert into FTS: %w", err)
		}
	}

	return tx.Commit()
}

func (idx *Index) DeleteNote(ctx context.Context, path string) error {
	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var noteID int64
	err = tx.QueryRowContext(ctx,
		"SELECT id FROM notes WHERE path = ?",
		path,
	).Scan(&noteID)

	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to query note: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM notes_fts WHERE rowid = ?",
		noteID,
	); err != nil {
		return fmt.Errorf("failed to delete from FTS: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		"DELETE FROM notes WHERE id = ?",
		noteID,
	); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return tx.Commit()
}

func (idx *Index) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}

	query = escapeFTS5Query(query)

	rows, err := idx.db.QueryContext(ctx, `
		SELECT n.path, n.title,
				COALESCE(snippet(notes_fts, 1, '**', '**', '...', 32), '') AS snippet,
				bm25(notes_fts) AS rank
		FROM notes_fts
		JOIN notes n ON n.id = notes_fts.rowid
		WHERE notes_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Path, &r.Title, &r.Snippet, &r.Rank); err != nil {
			continue
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

func (idx *Index) SearchByTag(ctx context.Context, tag string, limit int) ([]SearchResult, error) {
	return idx.Search(ctx, "#"+tag, limit)
}

func (idx *Index) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var count int
	err := idx.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM notes",
	).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count notes: %w", err)
	}
	stats["total_notes"] = count

	var lastIndexed sql.NullInt64
	err = idx.db.QueryRowContext(ctx,
		"SELECT MAX(indexed_at) FROM notes",
	).Scan(&lastIndexed)
	if err != nil {
		return nil, fmt.Errorf("failed to get last indexed: %w", err)
	}
	if lastIndexed.Valid {
		stats["last_indexed"] = time.Unix(lastIndexed.Int64, 0)
	}

	return stats, nil
}

func (idx *Index) GetIndexedPaths(ctx context.Context) (map[string]int64, error) {
	rows, err := idx.db.QueryContext(ctx,
		"SELECT path, mod_time FROM notes",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexed paths: %w", err)
	}
	defer rows.Close()

	paths := make(map[string]int64)
	for rows.Next() {
		var path string
		var modTime int64
		if err := rows.Scan(&path, &modTime); err != nil {
			continue
		}
		paths[path] = modTime
	}

	return paths, rows.Err()
}

func escapeFTS5Query(query string) string {
	query = strings.ReplaceAll(query, "\"", "\"\"")
	return fmt.Sprintf("\"%s\"", query)
}

func GetIndexPath(baseDir string) string {
	return filepath.Join(baseDir, ".jotr", "search.db")
}
