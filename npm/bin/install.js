#!/usr/bin/env node

const https = require("https");
const fs = require("fs");
const path = require("path");
const os = require("os");
const { execSync } = require("child_process");

const VERSION = require("../package.json").version;
const REPO = "moltyverse/clawdpl";

/**
 * Get platform-specific binary name
 */
function getPlatformInfo() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows",
  };

  const archMap = {
    x64: "amd64",
    arm64: "arm64",
  };

  const osPlatform = platformMap[platform];
  const osArch = archMap[arch];

  if (!osPlatform || !osArch) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }

  const ext = platform === "win32" ? ".exe" : "";
  const binaryName = `clawdpl${ext}`;
  const archiveName = `clawdpl_${VERSION}_${osPlatform}_${osArch}.tar.gz`;

  return { platform: osPlatform, arch: osArch, binaryName, archiveName };
}

/**
 * Download and extract the binary
 */
async function downloadBinary() {
  const { binaryName, archiveName } = getPlatformInfo();
  const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;
  const binDir = path.join(__dirname);
  const binaryPath = path.join(binDir, binaryName);

  // Skip if binary already exists
  if (fs.existsSync(binaryPath)) {
    console.log("clawdpl binary already installed.");
    return;
  }

  console.log(`Downloading clawdpl v${VERSION}...`);
  console.log(`URL: ${downloadUrl}`);

  try {
    // For now, just create a placeholder message since releases don't exist yet
    console.log("");
    console.log("Note: Pre-built binaries are not yet available.");
    console.log("Please install from source:");
    console.log("  go install github.com/moltyverse/clawdpl@latest");
    console.log("");
    console.log("Or build locally:");
    console.log("  git clone https://github.com/moltyverse/clawdpl.git");
    console.log("  cd clawdpl");
    console.log("  go build -o clawdpl .");
    console.log("");
  } catch (err) {
    console.error("Failed to download binary:", err.message);
    console.error("Please install from source: go install github.com/moltyverse/clawdpl@latest");
  }
}

downloadBinary().catch((err) => {
  console.error(err);
  process.exit(1);
});
