# jotr — Agent Guidelines

## Branch naming

All branches must follow: `as/jotr-XXXX-<slug>`

- Find the highest existing number with `git branch -a | grep -oP 'jotr-\d+' | sort -t- -k2 -n | tail -1`
- Increment by 1 for the new branch
- Slug should describe the change, no abbreviations

## Before every commit

Run `gofmt` and fix any files it flags before staging:

```bash
gofmt -l .          # list files needing formatting
gofmt -w <file>     # fix a specific file
```

A commit with unformatted Go files will break CI (`make test` runs `gofmt` checks).

## Build & test

```bash
make build   # must pass before committing
make test    # must pass before committing
```

## Language

This is a Go project. No TypeScript, no scripts unless already present.
