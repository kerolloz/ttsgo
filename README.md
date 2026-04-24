<p align="center">
  <strong>nego</strong> ⚡
</p>

<p align="center">
  Drop-in replacement for <code>nest build</code> and <code>nest start</code>.<br/>
  Powered by <a href="https://devblogs.microsoft.com/typescript/announcing-typescript-7-0-beta/">tsgo</a> — the native Go port of the TypeScript compiler.
</p>

<p align="center">
  <a href="#installation">Install</a> ·
  <a href="#usage">Usage</a> ·
  <a href="#how-it-works">How It Works</a> ·
  <a href="#contributing">Contributing</a>
</p>

---

## The Pitch

| Tool                    | Typical Build Time |
| ----------------------- | ------------------ |
| `nest build` (tsc)      | ~8 – 25 s          |
| **`nego build`** (tsgo) | **~0.5 – 2 s**     |

**Zero config changes.** `nego` reads your existing `nest-cli.json` and `tsconfig.json`. Drop the binary into any NestJS project and go.

---

## Requirements

| Dependency                      | Version                       |
| ------------------------------- | ----------------------------- |
| **Go** (build from source only) | ≥ 1.22                        |
| **Node.js**                     | ≥ 20                          |
| **tsgo**                        | installed locally or globally |

Install tsgo into your project:

```bash
npm install -D @typescript/native-preview
```

Or globally:

```bash
npm install -g @typescript/native-preview
```

`nego` will resolve the binary automatically — first from `node_modules/.bin/tsgo` (walking up the directory tree), then from `$PATH`.

---

## Installation

### Pre-built binary

Download the latest release from the [Releases](https://github.com/AhmedHafez13/nego/releases) page and place it on your `PATH`.

### Build from source

```bash
git clone https://github.com/AhmedHafez13/nego.git
cd nego
go build -o nego .
```

Or install directly:

```bash
go install github.com/AhmedHafez13/nego@latest
```

---

## Usage

### `nego build`

```bash
# One-shot build
nego build

# Watch mode — rebuild on every .ts file change
nego build --watch

# Custom tsconfig
nego build --path tsconfig.prod.json

# Custom nest-cli.json location
nego build --config nest-cli.prod.json

# Watch source AND asset files
nego build --watch --watchAssets
```

#### Build Flags

| Flag            | Short | Default     | Description                           |
| --------------- | ----- | ----------- | ------------------------------------- |
| `--config`      | `-c`  | auto-detect | Path to `nest-cli.json`               |
| `--path`        | `-p`  | auto-detect | Path to tsconfig file                 |
| `--watch`       | `-w`  | `false`     | Rebuild on `.ts` file changes         |
| `--watchAssets` |       | `false`     | Also watch non-TypeScript asset files |

---

### `nego start`

```bash
# Build and run
nego start

# Watch mode with hot-reload
nego start --watch

# Debug mode
nego start --debug
nego start --debug 0.0.0.0:9229
```

#### Start Flags

| Flag            | Short | Default     | Description                                  |
| --------------- | ----- | ----------- | -------------------------------------------- |
| `--config`      | `-c`  | auto-detect | Path to `nest-cli.json`                      |
| `--path`        | `-p`  | auto-detect | Path to tsconfig file                        |
| `--watch`       | `-w`  | `false`     | Live-reload on source changes                |
| `--watchAssets` |       | `false`     | Watch non-TypeScript asset files             |
| `--debug`       | `-d`  | off         | Enable `--inspect` (optionally `=host:port`) |
| `--exec`        | `-e`  | `node`      | Binary to run                                |
| `--entryFile`   |       | from config | Entry file name (no `.js` extension)         |
| `--sourceRoot`  |       | from config | Source root directory                        |
| `--no-shell`    |       | `false`     | Do not wrap the child process in a shell     |
| `--env-file`    |       |             | Path to a `.env` file (repeatable)           |

---

## How It Works

```
nego build
   │
   ├─ 1. Load nest-cli.json ──▶ sourceRoot, entryFile, assets, deleteOutDir
   ├─ 2. Load tsconfig.json ──▶ outDir, rootDir, paths, decorator flags
   │      └─ Follows the full "extends" chain (including node_modules)
   │      └─ Validates emitDecoratorMetadata + experimentalDecorators
   ├─ 3. Locate tsgo binary ──▶ node_modules/.bin/tsgo → $PATH
   ├─ 4. Delete outDir ───────▶ (if deleteOutDir: true)
   ├─ 5. Run tsgo -p <tsconfig>
   ├─ 6. Rewrite path aliases ▶ Parallel walk of dist/**/*.js
   └─ 7. Copy assets ─────────▶ Glob → dist/, preserving structure

nego start = build + spawn(node dist/main.js)
   └─ --watch: fsnotify on src/ → kill old process → rebuild → spawn new
```

### Watch Mode Internals

1. **File watching** — `fsnotify` watches the entire `sourceRoot` tree. New directories created at runtime (e.g., `src/new-module/`) are detected and added automatically.
2. **Debounce** — A 300 ms debounce window coalesces rapid file-save events into a single rebuild.
3. **Serialization** — A buffered channel of size 1 acts as a single-slot queue. Only one rebuild runs at a time; rapid changes during a build are coalesced into one pending rebuild.
4. **Process lifecycle** — Before each rebuild, the previous Node.js process tree is killed via `SIGTERM` to the process group (`-PGID`), with automatic escalation to `SIGKILL` after a 2-second grace period. No zombie processes.
5. **Graceful shutdown** — `Ctrl+C` triggers context cancellation, which cleanly stops all watchers, the rebuild loop, and the child process. A second `Ctrl+C` force-exits immediately.

### Path Alias Rewriting

`nego` has built-in support for TypeScript `paths`:

```jsonc
// tsconfig.json
{
  "compilerOptions": {
    "paths": {
      "~": ["./src"],
      "~/*": ["./src/*"],
      "@app/*": ["./src/app/*"],
    },
  },
}
```

After tsgo emits `.js` files, a fast parallel pass rewrites every `require()` / `import` / `export ... from` specifier in-place:

```js
// Before (emitted by tsgo — aliases left as-is)
const service = require("~/service");

// After (resolved to a correct relative path for Node.js)
const service = require("./service");
```

This mirrors what the official `nest-cli` does via its `tsconfigPathsBeforeHookFactory` AST transformer. The rewriter caches `node_modules` lookups and uses `runtime.NumCPU()` goroutines for parallel I/O.

### Decorator Metadata Validation

NestJS depends on `reflect-metadata` for its dependency injection container. At startup, `nego` reads the resolved tsconfig chain and warns if either of the following flags is missing:

- `emitDecoratorMetadata`
- `experimentalDecorators`

This catches a common misconfiguration that would otherwise surface as a cryptic runtime error.

---

## Supported `nest-cli.json` Fields

```jsonc
{
  "sourceRoot": "src",
  "entryFile": "main",
  "exec": "node",
  "compilerOptions": {
    "tsConfigPath": "tsconfig.build.json",
    "deleteOutDir": true,
    "assets": [
      "**/*.graphql",
      {
        "include": "**/*.proto",
        "exclude": "**/ignored",
        "outDir": "dist",
        "watchAssets": true,
      },
    ],
    "watchAssets": false,
  },
}
```

### Not Supported (Out of Scope)

| Feature                                 | Reason                                                         |
| --------------------------------------- | -------------------------------------------------------------- |
| `builder: "swc"` / `builder: "webpack"` | tsgo is the compiler — this tool replaces the tsc builder only |
| `compilerOptions.plugins`               | No TypeScript AST transformer plugin support                   |
| `projects` (monorepo mode)              | Single-project only                                            |
| `generate`, `new`, `info`, `add`        | Out of scope — use `nest` CLI for those                        |

---

## Architecture

```
nego/
├── main.go                           # CLI entry point (cobra)
├── cmd/
│   ├── build.go                      # `build` command — flag parsing, signal setup
│   └── start.go                      # `start` command — flag parsing, process runner setup
└── internal/
    ├── orchestrator/orchestrator.go   # Central build/watch pipeline
    ├── compiler/tsgo.go              # tsgo binary discovery and execution
    ├── config/
    │   ├── nestconfig.go             # nest-cli.json parser (union-typed assets, builder resolution)
    │   └── tsconfig.go               # tsconfig.json parser (JSONC, extends chain, node_modules)
    ├── paths/rewriter.go             # Post-emit path alias rewriter (parallel, cached)
    ├── assets/manager.go             # Asset glob/copy/watch manager
    ├── process/runner.go             # Node.js process lifecycle (PGID, SIGKILL escalation)
    └── watcher/watcher.go            # fsnotify-based file watcher (debounced, auto-adds dirs)
```

### Key Design Decisions

- **Interfaces for testability.** `compiler.Compiler` and `process.ProcessRunner` are interfaces. The orchestrator depends on abstractions, not concrete types — making it straightforward to unit test the build pipeline with mocks.
- **Context-driven lifecycle.** Every goroutine — the rebuild loop, file watchers, asset watchers — accepts a `context.Context` and exits cleanly on cancellation. No goroutine leaks.
- **Process group isolation.** Child processes are spawned with `Setpgid: true`. On kill, the entire process group is targeted (`syscall.Kill(-pgid, ...)`), preventing orphaned Node.js workers or database connection handlers.
- **Single `cwd` resolution.** The working directory is resolved once at the command level and threaded through every subsystem as an explicit parameter.

---

## Comparison with `nest build` / `nest start`

| Feature                       | `nest` CLI         | `nego`                          |
| ----------------------------- | ------------------ | ------------------------------- |
| Compiler                      | tsc                | **tsgo (~10× faster)**          |
| Path alias rewriting          | ✅ AST transformer | ✅ Post-emit parallel regex     |
| Asset copying                 | ✅                 | ✅                              |
| Asset watching                | ✅                 | ✅                              |
| `deleteOutDir`                | ✅                 | ✅                              |
| Watch mode                    | ✅ `tsc --watch`   | ✅ `fsnotify` + debounce        |
| Decorator metadata validation | ❌                 | ✅ Warns on missing flags       |
| Graceful shutdown             | Partial            | ✅ SIGTERM → SIGKILL escalation |
| webpack / swc builders        | ✅                 | ❌                              |
| TypeScript plugins            | ✅                 | ❌                              |
| Monorepo (`--all`)            | ✅                 | ❌                              |
| `nest generate`               | ✅                 | ❌                              |

---

## Contributing

Contributions are welcome. Please open an issue first to discuss what you'd like to change.

```bash
# Run all tests
go test ./... -v

# Lint
go vet ./...

# Build
go build -o nego .
```

### Running Locally Against a NestJS Project

```bash
# From the nego repo
go build -o nego .

# From your NestJS project
/path/to/nego start --watch
```

---

## License

MIT
