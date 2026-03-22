# AGENTS.md

## Project Summary

`texvec` is a small Go CLI for local text similarity search. It summarizes text documents with ONNX models, embeds summaries and document chunks, stores vectors in a local libsql database, and ranks matches with cosine distance.

## Repository Map

- `main.go`: CLI entrypoint.
- `cmd/`: Cobra command wiring. Keep these files thin and focused on argument parsing, orchestration, and user-facing output.
- `core/`: model registries, ONNX runtime setup, model downloads, text loading, chunking, summarization, embedding, and normalization.
- `store/`: database access, schema migration, inserts, listing, and vector search queries.
- `config/`: config bootstrapping and `~/.texvec` path helpers.
- `test_texts/`: sample documents for manual testing.

## Working Rules

- Prefer small, direct changes that match the current architecture.
- Keep reusable logic out of `cmd/`. Put business logic in `core/` or `store/`.
- Preserve existing CLI names and defaults unless the change explicitly requires a breaking UX change.
- Prefer the standard library and existing dependencies before adding new ones.
- If you add or change a model, update the model registry, user-facing help, and both READMEs together.
- Keep `summarize` preview-only and `embed` as the main indexing command unless the product direction explicitly changes.

## Build And Test

- Format Go changes with `gofmt -w`.
- Fast validation matches the most important packages: `go test ./store/ ./core/ -v`
- Build the CLI with `go build -o texvec`
- If you change package boundaries or command wiring, also run `go test ./...` when practical.

## Testing Guidance

- Prefer unit tests in `core/` and `store/`.
- Keep tests deterministic and offline.
- Do not make unit tests depend on model downloads, ONNX runtime downloads, or external services.
- Use temp directories and temp databases in tests rather than the real home directory.
- If a feature needs path-based testing, add an injection seam first instead of writing to `~/.texvec` in tests.

## Side Effects And Safety

- Runtime data lives outside the repo in `~/.texvec/`.
- `init`, `set-embedding-model`, `set-summary-model`, `embed`, and `search` may create directories or download runtime/model assets.
- `clean` is destructive because it removes all `texvec` data under `~/.texvec/`.
- Avoid adding tests that write to the real home directory.

## Documentation Expectations

- Update `README.md` for any user-visible command, flag, model, storage, platform, or workflow change.
- Keep `README.ja.md` reasonably aligned with the English README.
- Keep command examples aligned with actual Cobra behavior and current defaults.

## Useful Commands

```sh
go test ./store/ ./core/ -v
go test ./...
go build -o texvec
./texvec init
./texvec summarize test_texts/galaxies.md
./texvec embed test_texts/galaxies.md
./texvec search --text "dark matter in spiral galaxies"
```
