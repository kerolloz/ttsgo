# ttsgo & nestgo

**nestgo** is a drop-in replacement for `nest build` and `nest start`, powered by **ttsgo** — a TypeScript compiler built on top of TypeScript 7.0's native Go port.

No config changes required. Same `nest-cli.json`. Same `tsconfig.json`. Just faster.

---

## Benchmarks

Benchmarks were run on dummy projects. Real-world speedups depend on project size, but typically land between **5x and 30x**.

| Project      | Tool           | Time    | Speedup     |
|--------------|----------------|---------|-------------|
| TypeScript   | `tsc`          | 0.643s  | baseline    |
| TypeScript   | `ttsgo`        | 0.130s  | **~5x**     |
| NestJS app   | `nest build`   | 1.856s  | baseline    |
| NestJS app   | `nestgo build` | 0.060s  | **~30x**    |

---

## nestgo

A CLI replacement for the NestJS build toolchain. Reads your existing `nest-cli.json` and compiles with `ttsgo` instead of `tsc`.

### Installation

Install as a dev dependency and update your npm scripts:

```bash
npm install --save-dev nestgo
```

```json
// package.json
{
  "scripts": {
    "build":     "nestgo build",
    "start":     "nestgo start",
    "start:dev": "nestgo start --watch"
  }
}
```

### Usage

```bash
nestgo build                  # replaces: nest build
nestgo build --watch          # watch mode
nestgo start                  # replaces: nest start
nestgo start --watch          # hot-reload dev server
nestgo start --debug          # attach Node.js inspector
nestgo start -- --port 3001   # pass extra args to Node
```

All commands auto-detect `nest-cli.json` and `tsconfig.json` in the current directory. No flags required for standard project layouts.

### Supported `nest-cli.json` fields

| Field                             | Supported |
|-----------------------------------|-----------|
| `sourceRoot`                      | ✅        |
| `entryFile`                       | ✅        |
| `exec`                            | ✅        |
| `compilerOptions.tsConfigPath`    | ✅        |
| `compilerOptions.deleteOutDir`    | ✅        |
| `compilerOptions.assets`          | ✅        |
| `compilerOptions.watchAssets`     | ✅        |
| `compilerOptions.builder: "swc"`  | ❌        |
| `compilerOptions.builder: "webpack"` | ❌     |
| plugins, generate, add, new       | ❌        |

Only the `tsc` builder is supported. If your project uses `swc` or `webpack`, nestgo will exit with a clear error rather than silently producing wrong output.

### Path aliases

If your `tsconfig.json` defines `paths`, nestgo rewrites them in the emitted output automatically — no `tsc-alias` or post-processing step needed.

```json
// tsconfig.json
{
  "compilerOptions": {
    "paths": {
      "@modules/*": ["./src/modules/*"],
      "@common/*":  ["./src/common/*"]
    }
  }
}
```

This works for both `import ... from` and `require(...)` syntax, including `.d.ts` declaration files.

### Environment variables

| Variable            | Description                                          |
|---------------------|------------------------------------------------------|
| `NESTGO_DEBOUNCE_MS`| File watcher debounce in milliseconds (default: 500) |
| `TTSGO_WORKERS`     | Number of I/O workers for parallel file emit         |

---

## ttsgo

A standalone TypeScript compiler for non-NestJS projects. A faster drop-in for `tsc`.

### Installation

```bash
npm install --save-dev ttsgo
```

### Usage

```bash
ttsgo -p tsconfig.json           # compile project
ttsgo -p tsconfig.json --noEmit  # type-check only, no output
ttsgo -p tsconfig.json --outDir dist
```

`ttsgo` reads your `tsconfig.json`, runs type-checking and emit using TypeScript 7.0's native Go compiler, and rewrites any `paths` aliases in the emitted files in one pass — no separate post-processing step.

---

## How it works

```
nestgo build
     │
     ├── reads nest-cli.json + tsconfig.json
     │
     └── calls ttsgo engine (in-process, no child process)
              │
              ├── TypeScript 7.0 Go compiler (microsoft/typescript-go)
              │     type-check + emit
              │
              ├── concurrent I/O worker pool
              │     parallel WriteFile across CPU cores
              │
              └── paths rewriter
                    rewrites @aliases → relative paths
                    in-memory, during emit
```

The compiler runs in the same process as nestgo — there's no `exec.Command` spawned for compilation. The only child process is the Node.js app itself (`nestgo start`).

Path alias rewriting happens as each file buffer is handed off to disk, so there's no second pass over the output directory.

---

## Local development

Requires [just](https://github.com/casey/just), [Go 1.26+](https://github.com/kerolloz/go-installer) and Node.js 20+.

```bash
git clone https://github.com/kerolloz/ttsgo.git
cd ttsgo
just build          # builds both binaries to ./bin/
```

Or manually:

```bash
# build ttsgo
cd packages/ttsgo
go build -o ../../bin/ttsgo ./cmd/ttsgo

# build nestgo
cd packages/nestgo
go build -o ../../bin/nestgo .
```

Run against the included test project:

```bash
cd tests/dummy-ts
../../bin/ttsgo -p tsconfig.json
```

---

## License

MIT
