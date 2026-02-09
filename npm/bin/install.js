#!/usr/bin/env node

const https = require("https");
const fs = require("fs");
const path = require("path");
const os = require("os");
const { execSync, spawn } = require("child_process");
const zlib = require("zlib");

const VERSION = require("../package.json").version;
const REPO = "moltyverse/clawdepl";

/**
 * Get platform-specific binary info
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
  const binaryName = `clawdepl${ext}`;
  const archiveName = `clawdepl_${VERSION}_${osPlatform}_${osArch}.tar.gz`;

  return { platform: osPlatform, arch: osArch, binaryName, archiveName, ext };
}

/**
 * Follow redirects and download file
 */
function downloadFile(url) {
  return new Promise((resolve, reject) => {
    const makeRequest = (currentUrl, redirectCount = 0) => {
      if (redirectCount > 10) {
        reject(new Error("Too many redirects"));
        return;
      }

      const protocol = currentUrl.startsWith("https") ? https : require("http");
      
      protocol.get(currentUrl, {
        headers: {
          "User-Agent": "clawdepl-npm-installer",
          "Accept": "application/octet-stream",
        }
      }, (response) => {
        if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
          makeRequest(response.headers.location, redirectCount + 1);
          return;
        }

        if (response.statusCode !== 200) {
          reject(new Error(`Download failed with status ${response.statusCode}`));
          return;
        }

        const chunks = [];
        response.on("data", (chunk) => chunks.push(chunk));
        response.on("end", () => resolve(Buffer.concat(chunks)));
        response.on("error", reject);
      }).on("error", reject);
    };

    makeRequest(url);
  });
}

/**
 * Extract tar.gz archive
 */
async function extractTarGz(buffer, destDir, binaryName) {
  const tar = require("tar");
  const tmpFile = path.join(os.tmpdir(), `clawdepl-${Date.now()}.tar.gz`);
  
  fs.writeFileSync(tmpFile, buffer);
  
  try {
    await tar.extract({
      file: tmpFile,
      cwd: destDir,
      filter: (p) => p === binaryName || p.endsWith(`/${binaryName}`),
    });
  } finally {
    fs.unlinkSync(tmpFile);
  }
}

/**
 * Simple tar extraction without external dependency
 */
function extractTarGzSimple(buffer, destDir, binaryName) {
  return new Promise((resolve, reject) => {
    const gunzip = zlib.createGunzip();
    const chunks = [];
    
    gunzip.on("data", (chunk) => chunks.push(chunk));
    gunzip.on("end", () => {
      const tarData = Buffer.concat(chunks);
      
      // Simple tar parsing - find the binary file
      let offset = 0;
      while (offset < tarData.length) {
        // Read header (512 bytes)
        const header = tarData.slice(offset, offset + 512);
        if (header[0] === 0) break; // End of archive
        
        // Get filename (first 100 bytes, null-terminated)
        let nameEnd = 0;
        while (nameEnd < 100 && header[nameEnd] !== 0) nameEnd++;
        const name = header.slice(0, nameEnd).toString("utf8");
        
        // Get file size (octal, bytes 124-135)
        const sizeStr = header.slice(124, 136).toString("utf8").trim();
        const size = parseInt(sizeStr, 8) || 0;
        
        offset += 512; // Move past header
        
        // Check if this is our binary
        if (name === binaryName || name.endsWith(`/${binaryName}`)) {
          const fileData = tarData.slice(offset, offset + size);
          const destPath = path.join(destDir, binaryName);
          fs.writeFileSync(destPath, fileData);
          fs.chmodSync(destPath, 0o755);
          resolve(destPath);
          return;
        }
        
        // Move to next file (size rounded up to 512-byte boundary)
        offset += Math.ceil(size / 512) * 512;
      }
      
      reject(new Error(`Binary ${binaryName} not found in archive`));
    });
    gunzip.on("error", reject);
    
    gunzip.end(buffer);
  });
}

/**
 * Download and install the binary
 */
async function downloadBinary() {
  const { binaryName, archiveName } = getPlatformInfo();
  const binDir = __dirname;
  const binaryPath = path.join(binDir, binaryName);

  // Skip if binary already exists
  if (fs.existsSync(binaryPath)) {
    console.log("clawdepl binary already installed.");
    return;
  }

  // Check for local development binary (in project root)
  const devBinary = path.join(__dirname, "..", "..", binaryName);
  if (fs.existsSync(devBinary)) {
    console.log("Using local development binary.");
    fs.copyFileSync(devBinary, binaryPath);
    fs.chmodSync(binaryPath, 0o755);
    return;
  }

  const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;
  
  console.log(`Downloading clawdepl v${VERSION}...`);
  console.log(`Platform: ${os.platform()}-${os.arch()}`);
  console.log(`URL: ${downloadUrl}`);

  try {
    const buffer = await downloadFile(downloadUrl);
    console.log(`Downloaded ${buffer.length} bytes, extracting...`);
    
    await extractTarGzSimple(buffer, binDir, binaryName);
    
    console.log(`Successfully installed clawdepl to ${binaryPath}`);
  } catch (err) {
    console.warn("");
    console.warn(`Note: Could not download pre-built binary: ${err.message}`);
    console.warn("");
    console.warn("This is expected if:");
    console.warn("  - This version hasn't been released yet");
    console.warn("  - You're on an unsupported platform");
    console.warn("");
    console.warn("The CLI will look for 'clawdepl' in your PATH at runtime.");
    console.warn("Install from source: go install github.com/moltyverse/clawdepl@latest");
    console.warn("");
    // Don't fail the install - the wrapper will handle missing binary at runtime
  }
}

downloadBinary().catch((err) => {
  console.error("Installation warning:", err.message);
  // Exit 0 to not fail npm install - wrapper handles missing binary gracefully
  process.exit(0);
});
