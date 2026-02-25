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

	// SQLite requires single connection for WAL mode and to avoid "database is locked" errors
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
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY
		);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create base schema: %w", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count); err != nil {
		return fmt.Errorf("failed to count schema version: %w", err)
	}
	if count == 0 {
		if _, err := db.Exec("INSERT INTO schema_version (version) VALUES (0)"); err != nil {
			return fmt.Errorf("failed to initialize schema version: %w", err)
		}
	}

	var version int
	if err := db.QueryRow("SELECT version FROM schema_version ORDER BY version DESC LIMIT 1").Scan(&version); err != nil {
		return fmt.Errorf("failed to read schema version: %w", err)
	}

	if version < 1 {
		// V1: Initial FTS5 table with contentless_delete=1
		migrationV1 := `
			DROP TABLE IF EXISTS notes_fts;
			DROP TABLE IF EXISTS notes;
            
			CREATE TABLE notes (
				id INTEGER PRIMARY KEY,
				path TEXT UNIQUE NOT NULL,
				title TEXT,
				mod_time INTEGER NOT NULL,
				indexed_at INTEGER NOT NULL
			);

			CREATE INDEX idx_notes_path ON notes(path);
			CREATE INDEX idx_notes_mod_time ON notes(mod_time);

			CREATE VIRTUAL TABLE notes_fts USING fts5(
				title,
				content,
				content='',
				contentless_delete=1,
				tokenize="trigram"
			);
			UPDATE schema_version SET version = 1;
		`
		if _, err := db.Exec(migrationV1); err != nil {
			return fmt.Errorf("failed to apply migration v1: %w", err)
		}
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
				COALESCE(snippet(notes_fts, 1, '>>>', '<<<', '...', 32), '') AS snippet,
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
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return results, nil
}

func (idx *Index) SearchByTag(ctx context.Context, tag string, limit int) ([]SearchResult, error) {
	candidateLimit := limit
	if candidateLimit == 0 {
		candidateLimit = 500
	}
	candidates, err := idx.Search(ctx, "#"+tag, candidateLimit)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	exactTag := "#" + tag

	for _, c := range candidates {
		content, err := os.ReadFile(c.Path)
		if err != nil {
			continue
		}

		if hasExactTag(string(content), exactTag) {
			results = append(results, c)
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

func hasExactTag(content, exactTag string) bool {
	idx := 0
	tagLen := len(exactTag)
	contentLen := len(content)

	for {
		i := strings.Index(content[idx:], exactTag)
		if i == -1 {
			return false
		}

		pos := idx + i
		afterIdx := pos + tagLen
		if afterIdx < contentLen {
			nextChar := content[afterIdx]
			if (nextChar >= 'a' && nextChar <= 'z') ||
				(nextChar >= 'A' && nextChar <= 'Z') ||
				(nextChar >= '0' && nextChar <= '9') ||
				nextChar == '_' || nextChar == '-' {
				idx = pos + 1
				continue
			}
		}

		if pos > 0 {
			prevChar := content[pos-1]
			if (prevChar >= 'a' && prevChar <= 'z') ||
				(prevChar >= 'A' && prevChar <= 'Z') ||
				(prevChar >= '0' && prevChar <= '9') {
				idx = pos + 1
				continue
			}
		}

		return true
	}
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
			return nil, fmt.Errorf("failed to scan indexed path: %w", err)
		}
		paths[path] = modTime
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return paths, nil
}

func escapeFTS5Query(query string) string {
	replacer := strings.NewReplacer(
		`"`, `""`,
		`*`, ``,
		`^`, ``,
		`(`, ``,
		`)`, ``,
	)
	escaped := replacer.Replace(query)
	return fmt.Sprintf(`"%s"`, escaped)
}

func GetIndexPath(baseDir string) string {
	return filepath.Join(baseDir, ".jotr", "search.db")
}
