# Contributing

Thanks for contributing to `texvec`.

## Before opening a PR

- Keep CLI behavior and defaults stable unless the change intentionally updates user-facing UX.
- Keep reusable logic out of `cmd/`. Put business logic in `core/` or `store/`.
- Update `README.md` for user-visible changes. Update `README.ja.md` too when practical.
- Prefer the standard library and existing dependencies before adding new ones.
- If a change touches model behavior, update the model registry, help text, and docs together.

## Development

```sh
go test ./...
go build -o texvec
```

## Testing

- Prefer deterministic unit tests in `core/` and `store/`.
- Do not make unit tests depend on model downloads, ONNX Runtime downloads, or external services.
- Use temp directories and temp databases instead of writing to the real home directory.
- Keep smoke tests separate from unit tests. Manual validation with `test_texts/` is fine, but it should not be required for CI.

## Pull Requests

- Keep changes focused and explain the user-facing impact.
- Add or update tests when behavior changes.
- Update docs and examples when commands, flags, models, storage, or workflows change.
- Call out any new model downloads, runtime requirements, or compatibility assumptions.

## AI Coding Agents

If you use an AI coding agent, read [AGENTS.md](AGENTS.md) before making changes.

## License

By submitting a contribution, you agree that your work will be licensed under the MIT License.
