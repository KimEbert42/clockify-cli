# clockify-cli (npm)

npm wrapper that installs the native `clockify` binary from GitHub Releases.

## Install

```bash
npm install -g clockify-cli
clockify version
```

If the unscoped name `clockify-cli` is taken on the public registry, publish under a scope:

```bash
npm install -g @kim/clockify-cli
```

## Install a specific version

```bash
CLOCKIFY_VERSION=0.10.0 npm install -g clockify-cli
```

## Notes

- `postinstall` downloads the matching platform binary (linux/darwin/windows × amd64/arm64).
- If the postinstall download fails (offline, private network), install manually from GitHub Releases or `go install github.com/kim/clockify-cli/cmd/clockify@latest`.
- Binary is placed at `node_modules/clockify-cli/bin/clockify` and linked as the `clockify` bin.
