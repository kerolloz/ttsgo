#!/usr/bin/env node
"use strict";

// Launcher script for the nego binary.
// Locates the platform-specific binary from @nego/core-<platform> and
// executes it with the same arguments. Follows the esbuild/ttsc pattern.

const { execFileSync } = require("child_process");
const { existsSync, statSync, chmodSync } = require("fs");
const path = require("path");

const PLATFORMS = {
  "darwin-arm64": { pkg: "@nego/core-darwin-arm64", bin: "nego" },
  "darwin-x64":   { pkg: "@nego/core-darwin-x64",   bin: "nego" },
  "linux-x64":    { pkg: "@nego/core-linux-x64",    bin: "nego" },
  "linux-arm64":  { pkg: "@nego/core-linux-arm64",  bin: "nego" },
  "win32-x64":    { pkg: "@nego/core-win32-x64",    bin: "nego.exe" },
  "win32-arm64":  { pkg: "@nego/core-win32-arm64",  bin: "nego.exe" },
};

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
  if (process.env.NEGO_BINARY) {
    return process.env.NEGO_BINARY;
  }

  const localBinary = path.join(__dirname, "..", "..", "..", "bin", "nego");
  if (existsSync(localBinary)) {
    return localBinary;
  }

  const key = `${process.platform}-${process.arch}`;
  const entry = PLATFORMS[key];
  if (!entry) {
    console.error(
      `nego: unsupported platform ${key}\n` +
      `No pre-built binary is available for this platform.`
    );
    process.exit(1);
  }

  try {
    return require.resolve(`${entry.pkg}/bin/${entry.bin}`);
  } catch {
    console.error(
      `nego: could not find binary package ${entry.pkg}\n` +
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
  console.error(`nego: failed to execute ${binPath}`);
  console.error(err.message);
  process.exit(1);
}
