#!/usr/bin/env node
"use strict";

// Launcher script for the ttsgo binary.
// Locates the platform-specific binary from @ttsgo/core-<platform> and
// executes it with the same arguments. Follows the esbuild/ttsc pattern.
//
// macOS Gatekeeper: Binaries installed via npm do NOT receive the
// com.apple.quarantine extended attribute, so Gatekeeper never blocks them.
// This is why the npm distribution channel works seamlessly on macOS.

const { execFileSync } = require("child_process");
const { existsSync, statSync, chmodSync } = require("fs");
const path = require("path");

const PLATFORMS = {
  "darwin-arm64": { pkg: "@ttsgo/core-darwin-arm64", bin: "ttsgo" },
  "darwin-x64":   { pkg: "@ttsgo/core-darwin-x64",   bin: "ttsgo" },
  "linux-x64":    { pkg: "@ttsgo/core-linux-x64",    bin: "ttsgo" },
  "linux-arm64":  { pkg: "@ttsgo/core-linux-arm64",  bin: "ttsgo" },
  "win32-x64":    { pkg: "@ttsgo/core-win32-x64",    bin: "ttsgo.exe" },
  "win32-arm64":  { pkg: "@ttsgo/core-win32-arm64",  bin: "ttsgo.exe" },
};

/**
 * Ensure the binary has the executable bit set.
 * npm doesn't always preserve file modes across all platforms.
 */
function ensureExecutable(binary) {
  if (process.platform === "win32") return;
  try {
    const mode = statSync(binary).mode & 0o777;
    if ((mode & 0o111) !== 0) return;
    chmodSync(binary, mode | 0o755);
  } catch {
    // keep the original spawn error path
  }
}

function getBinaryPath() {
  // 1. Environment variable override
  if (process.env.TTSGO_BINARY) {
    return process.env.TTSGO_BINARY;
  }

  // 2. Local development (monorepo)
  const localBinary = path.join(__dirname, "..", "..", "..", "bin", "ttsgo");
  if (existsSync(localBinary)) {
    return localBinary;
  }

  // 3. Resolve from platform-specific optional dependency
  const key = `${process.platform}-${process.arch}`;
  const entry = PLATFORMS[key];
  if (!entry) {
    console.error(
      `ttsgo: unsupported platform ${key}\n` +
      `No pre-built binary is available for this platform.`
    );
    process.exit(1);
  }

  try {
    return require.resolve(`${entry.pkg}/bin/${entry.bin}`);
  } catch {
    console.error(
      `ttsgo: could not find binary package ${entry.pkg}\n` +
      `Make sure it is installed. You may need to run "npm install".`
    );
    process.exit(1);
  }
}

const binPath = getBinaryPath();
ensureExecutable(binPath);

try {
  execFileSync(binPath, process.argv.slice(2), {
    stdio: "inherit",
    env: process.env,
  });
} catch (err) {
  if (err.status != null) {
    process.exit(err.status);
  }
  console.error(`ttsgo: failed to execute ${binPath}`);
  console.error(err.message);
  process.exit(1);
}
