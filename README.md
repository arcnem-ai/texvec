<p align="center">
  <img src="arcnem-logo.svg" alt="Arcnem AI" width="120" />
</p>

<h1 align="center">texvec</h1>

<p align="center">
  <strong>Local-first text similarity search.</strong>
</p>

<p align="center">
  <a href="README.ja.md">日本語</a> ·
  <a href="#install">Install</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#models">Models</a> ·
  <a href="#development">Development</a>
</p>

---

texvec is an open-source CLI for text similarity search. It summarizes documents locally with ONNX models, embeds both summaries and overlapping document chunks, stores the vectors in a local libsql database, and ranks matches with cosine distance.

Built by Arcnem AI, texvec reflects how we like to ship applied AI tools: local-first, inspectable, and useful without a cloud control plane.

## Why texvec

- Local summaries, local embeddings, local storage. Your documents stay on your machine.
- Simple CLI workflow. `init`, `summarize`, `embed`, `search`, and `list` are enough to get useful results quickly.
- No external vector database. Similarity search runs from a local libsql database.
- Practical text indexing. texvec stores both a document summary embedding and chunk embeddings so searches can match the gist or a specific section.

## Install

Download a release asset from GitHub Releases, or install with Go:

```sh
go install github.com/arcnem-ai/texvec@latest
```

To build from source:

```sh
git clone https://github.com/arcnem-ai/texvec.git
cd texvec
go build -o texvec
```

Release archives are expected to follow the same primary targets as `picvec`: macOS (`arm64`) and Linux (`amd64`).

## Quick Start

```sh
texvec init
texvec summarize test_texts/galaxies.md
texvec embed test_texts/galaxies.md
texvec search --text "dark matter in spiral galaxies"
texvec list
```

`texvec init` downloads ONNX Runtime, creates `~/.texvec/`, initializes the local database, and fetches the default summary and embedding models.

## Commands

| Command | What it does |
|---------|--------------|
| `init` | Download ONNX Runtime and the default models |
| `summarize [document]` | Generate and print a summary without writing to the database |
| `embed [document]` | Summarize, chunk, embed, and store a document |
| `search [document]` | Find similar indexed documents |
| `search --text "..."` | Search from raw text |
| `list` | List indexed documents |
| `set-embedding-model [name]` | Set the default embedding model |
| `set-summary-model [name]` | Set the default summary model |
| `config` | Show current configuration |
| `clean` | Remove all `texvec` data |

Global flag:

- `-v, --verbose` enables extra runtime output

### Common Examples

Preview a summary:

```sh
texvec summarize notes.md
texvec summarize notes.md --summary-model flan-t5-small
```

`texvec summarize` is preview-only. It does not write to the database.

Index a document:

```sh
texvec embed notes.md
texvec embed notes.md -m bge-small-en-v1.5
texvec embed notes.md --summary-model flan-t5-small
```

If the document content hash is unchanged, texvec reuses existing summary and chunk data where possible.

Search for similar documents:

```sh
texvec search notes.md
texvec search --text "barred spiral galaxy dark matter"
texvec search --text "barred spiral galaxy dark matter" -k 10
texvec search --text "barred spiral galaxy dark matter" -k 10 -c 3
texvec search notes.md -m bge-small-en-v1.5
```

Results are ranked by summary cosine distance. For each matching document, texvec also prints the top chunk matches as supporting evidence. When searching with an already indexed document path, texvec excludes that same path from the results.

| Flag | Description | Default |
|------|-------------|---------|
| `-k, --limit` | Number of results | 5 |
| `-c, --chunks` | Supporting chunks to show per result | 1 |
| `-m, --model` | Embedding model to use | Config default |
| `--summary-model` | Summary model to use for long-query reduction | Config default |

List indexed documents:

```sh
texvec list
texvec list -m all-minilm-l6-v2
texvec list -k 20
```

| Flag | Description | Default |
|------|-------------|---------|
| `-k, --limit` | Max documents to show | All |
| `-m, --model` | Filter by embedding model | All |

Change defaults:

```sh
texvec set-embedding-model bge-small-en-v1.5
texvec set-summary-model flan-t5-small
```

## Models

### Embedding Models

| Name | Embedding Dim | Notes |
|------|---------------|-------|
| `all-minilm-l6-v2` | 384 | Default. Fast and good for general-purpose retrieval. |
| `bge-small-en-v1.5` | 384 | Retrieval-focused model with a query prefix for search. |

### Summary Models

| Name | Notes |
|------|-------|
| `flan-t5-small` | Default summary model for `1.0.0`. Small, local, and easy to ship in a plain Go CLI. |

Models are downloaded from Hugging Face on first use and stored locally under `~/.texvec/models/`.

## How It Works

1. A supported text document is loaded from `.txt`, `.md`, or `.markdown`.
2. texvec computes a content hash to determine whether indexing work needs to be refreshed.
3. The selected summary model generates a document summary.
4. The selected embedding model embeds both the summary and overlapping chunks from the original document.
5. Search compares the query embedding against stored summary embeddings to rank documents.
6. For the top documents, texvec compares the same query against stored chunk embeddings and prints the best chunk matches as evidence.

## Data Storage

All runtime data lives in `~/.texvec/`:

```text
~/.texvec/
  config.json       # Configuration such as the default models
  texvec.db         # libsql database
  models/           # Downloaded ONNX model files and tokenizer assets
  lib/              # ONNX Runtime shared library
```

Use `texvec clean` to remove everything.

## Repository Layout

- `cmd/` Cobra commands and user-facing output
- `core/` Model registry, runtime setup, downloads, text chunking, summarization, and embedding pipeline
- `store/` Schema migration, inserts, listing, and vector search queries
- `config/` `~/.texvec` path helpers and config bootstrapping
- `test_texts/` Sample documents for manual testing

## Platforms

| OS | Published Release | Hardware Acceleration |
|----|-------------------|----------------------|
| macOS | `arm64` | CPU |
| Linux | `amd64` | CPU |

texvec currently defaults to CPU execution for predictable CLI behavior and cleaner output across machines.

ONNX Runtime 1.24.3 is downloaded automatically on first run.

## Development

```sh
go test ./...
go build -o texvec
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution workflow and [AGENTS.md](AGENTS.md) for repo-specific agent instructions.

---

<p align="center">
  Built by <a href="https://arcnem.ai">Arcnem AI</a>.
</p>
