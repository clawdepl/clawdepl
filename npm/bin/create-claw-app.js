#!/usr/bin/env node

const { spawn } = require("child_process");
const path = require("path");
const os = require("os");
const fs = require("fs");

/**
 * Get the path to the create-claw-app binary for the current platform
 */
function getBinaryPath() {
  const platform = os.platform();
  const arch = os.arch();

  let binaryName = "create-claw-app";
  if (platform === "win32") {
    binaryName += ".exe";
  }

  // Check for binary in the package's bin directory
  const localBinary = path.join(__dirname, binaryName);
  if (fs.existsSync(localBinary)) {
    return localBinary;
  }

  // Check for binary in common installation locations
  const platformArch = `${platform}-${arch}`;
  const platformBinary = path.join(__dirname, "..", "bin", platformArch, binaryName);
  if (fs.existsSync(platformBinary)) {
    return platformBinary;
  }

  // Fall back to PATH
  return binaryName;
}

/**
 * Run the create-claw-app binary with the provided arguments
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
      console.error("Error: create-claw-app binary not found.");
      console.error("Please ensure the binary is installed correctly.");
      console.error("You can also install from source: go install github.com/moltyverse/create-claw-app@latest");
      process.exit(1);
    }
    console.error("Error running create-claw-app:", err.message);
    process.exit(1);
  });

  child.on("close", (code) => {
    process.exit(code ?? 0);
  });
}

run();
