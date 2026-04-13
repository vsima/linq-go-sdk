# Contributing

Thanks for your interest! A few ground rules:

## Workflow

1. Open an issue first for non-trivial changes so we can agree on scope.
2. Fork, branch from `main`, and make your change.
3. Run `make ci` locally before pushing.
4. Open a PR. Include a clear summary and tests for new behavior.

## Development

```sh
git clone https://github.com/vsima/linq-go-sdk
cd linq-go-sdk
make test
```

Code must:

- Pass `go vet ./...`
- Be formatted with `gofmt`
- Have no new third-party dependencies (stdlib only)
- Include tests for new exported surface

## Commit style

Keep commits focused. Conventional-style prefixes (`feat:`, `fix:`, `docs:`, `test:`, `chore:`) are nice but not required.

## Reporting security issues

Do **not** file public issues for security bugs. See [SECURITY.md](./SECURITY.md).
