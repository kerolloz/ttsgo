# ttsgo & nego

This repository contains two revolutionary tools designed to supercharge your TypeScript and NestJS development workflows by leveraging the speed of native Go.

*   [**`ttsgo`**](#ttsgo): An ultra-fast native TypeScript compiler engine.
*   [**`nego`**](#nego): A drop-in replacement for the NestJS CLI (`nest build` and `nest start`).

---

## Benchmarks

Both tools achieve phenomenal performance improvements by using native Go instead of V8 and Node.js.

| Project Type | Tool | Compilation Time | Speedup |
| :--- | :--- | :--- | :--- |
| Standard TS | `tsc` | 0.643s | Baseline |
| Standard TS | **`ttsgo`** | 0.130s | **~5x Faster** |
| NestJS App | `nest build`| 1.856s | Baseline |
| NestJS App | **`nego`** | 0.060s | **~30x Faster** |

> *Note: Benchmarks performed on dummy projects. Real-world speedups may vary, but typically sit between 5x and 30x.*

---

## 🚀 `ttsgo` (TypeScript Compiler)

`ttsgo` uses TypeScript 7.0's native Go compiler (`github.com/microsoft/typescript-go`) but extends it with **Zero-I/O Path Alias Resolution**. 

Instead of waiting for the compiler to emit `dist` files and then running a secondary tool like `tsc-alias` to replace `@utils/*` with relative paths, `ttsgo` rewrites aliases *in-memory* as the file buffer is written to disk.

### Features
- Native performance (up to 10x faster than `tsc`)
- Built-in `tsconfig.json` `paths` resolution. No `tsc-alias` needed!
- Produces exact same diagnostics as standard `tsc`.

### Usage
```bash
npm install -g ttsgo

# Compile your project using the tsconfig.json in the current directory
ttsgo -p tsconfig.json
```

---

## ⚡ `nego` (NestJS CLI Replacement)

`nego` is a lightweight, ultra-fast CLI orchestrator tailored specifically for NestJS applications. It completely bypasses Node.js compilation and utilizes `ttsgo`'s engine as a direct library binding (`//go:linkname`).

### Features
- In-process TypeScript compilation (no child process overhead).
- Understands `nest-cli.json` natively.
- Handles NestJS Asset compilation (e.g. `*.graphql`, `*.html`).
- Watch mode (`nego start --watch`) with lightning-fast rebuilds.

### Usage
```bash
npm install -g nego

# Build your NestJS app (replaces `nest build`)
nego build

# Build and start your NestJS app (replaces `nest start`)
nego start

# Watch mode
nego start --watch
```

---

## Architecture

This is a Go workspace Monorepo containing the following structures:

1. **`packages/ttsgo`**: Exposes both the CLI binary and the `pkg/engine` module. Uses `//go:linkname` shims to pierce the `internal/` barrier of `typescript-go`.
2. **`packages/nego`**: Consumes `ttsgo/pkg/engine` to orchestrate builds without executing `exec.Command`.

### Development

To build the binaries locally:
```bash
# Build ttsgo
cd packages/ttsgo
go build -o ../../bin/ttsgo ./cmd/ttsgo

# Build nego
cd packages/nego
go build -o ../../bin/nego ./main.go
```
