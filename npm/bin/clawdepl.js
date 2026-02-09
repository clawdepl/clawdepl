#!/usr/bin/env node

const { spawn } = require("child_process");
const path = require("path");
const os = require("os");
const fs = require("fs");

/**
 * Get the path to the clawdepl binary for the current platform
 */
function getBinaryPath() {
  const platform = os.platform();
  let binaryName = "clawdepl";
  if (platform === "win32") {
    binaryName += ".exe";
  }

  // 1. Check for binary in the same directory as this script (installed by postinstall)
  const localBinary = path.join(__dirname, binaryName);
  if (fs.existsSync(localBinary)) {
    return localBinary;
  }

  // 2. Check for development binary in project root
  const devBinary = path.join(__dirname, "..", "..", binaryName);
  if (fs.existsSync(devBinary)) {
    return devBinary;
  }

  // 3. Check for binary via CLAWDPL_BINARY_PATH env var (useful for testing)
  if (process.env.CLAWDPL_BINARY_PATH && fs.existsSync(process.env.CLAWDPL_BINARY_PATH)) {
    return process.env.CLAWDPL_BINARY_PATH;
  }

  // 4. Fall back to PATH lookup
  return binaryName;
}

/**
 * Run the clawdepl binary with the provided arguments
 */
function run() {
  const binaryPath = getBinaryPath();
  const args = process.argv.slice(2);

  const child = spawn(binaryPath, args, {
    stdio: "inherit",
    shell: process.platform === "win32",
  });

  child.on("error", (err) => {
    if (err.code === "ENOENT") {
      console.error("Error: clawdepl binary not found.");
      console.error("");
      console.error("The binary was not bundled with this package and is not in your PATH.");
      console.error("");
      console.error("Install the binary using one of these methods:");
      console.error("");
      console.error("  # Install from source (requires Go 1.21+):");
      console.error("  go install github.com/moltyverse/clawdepl@latest");
      console.error("");
      console.error("  # Or set CLAWDPL_BINARY_PATH to point to the binary:");
      console.error("  export CLAWDPL_BINARY_PATH=/path/to/clawdepl");
      console.error("");
      process.exit(1);
    }
    console.error("Error running clawdepl:", err.message);
    process.exit(1);
  });

  child.on("close", (code) => {
    process.exit(code ?? 0);
  });
}

run();
