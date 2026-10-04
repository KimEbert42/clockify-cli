#!/usr/bin/env node
/**
 * Downloads the clockify binary from GitHub Releases into bin/clockify.
 * Used as the npm postinstall step for the clockify-cli wrapper package.
 */
"use strict";

const fs = require("fs");
const path = require("path");
const https = require("https");
const { execSync } = require("child_process");
const { pipeline } = require("stream/promises");
const { createWriteStream } = require("fs");
const { createGunzip } = require("zlib");
const { Readable } = require("stream");

const REPO_OWNER = process.env.CLOCKIFY_REPO_OWNER || "kim";
const REPO_NAME = process.env.CLOCKIFY_REPO_NAME || "clockify-cli";
const VERSION = process.env.CLOCKIFY_VERSION || ""; // empty = latest release
const API_BASE = process.env.CLOCKIFY_API_BASE || "https://api.github.com";

function goos() {
  if (process.platform === "win32") return "windows";
  if (process.platform === "darwin") return "darwin";
  return "linux";
}

function goarch() {
  if (process.arch === "arm64") return "arm64";
  return "amd64";
}

function getJSON(url) {
  return new Promise((resolve, reject) => {
    https
      .get(
        url,
        {
          headers: {
            "User-Agent": "clockify-cli-npm",
            Accept: "application/vnd.github+json",
          },
        },
        (res) => {
          if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
            getJSON(res.headers.location).then(resolve, reject);
            return;
          }
          let data = "";
          res.on("data", (c) => (data += c));
          res.on("end", () => {
            if (res.statusCode !== 200) {
              reject(new Error(`GET ${url} -> ${res.statusCode}: ${data.slice(0, 200)}`));
              return;
            }
            try {
              resolve(JSON.parse(data));
            } catch (e) {
              reject(e);
            }
          });
        }
      )
      .on("error", reject);
  });
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    https
      .get(
        url,
        {
          headers: { "User-Agent": "clockify-cli-npm" },
        },
        (res) => {
          if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
            download(res.headers.location, dest).then(resolve, reject);
            return;
          }
          if (res.statusCode !== 200) {
            reject(new Error(`download ${url} -> ${res.statusCode}`));
            return;
          }
          const out = createWriteStream(dest);
          res.pipe(out);
          out.on("finish", () => out.close(resolve));
          out.on("error", reject);
        }
      )
      .on("error", reject);
  });
}

function pickAsset(assets) {
  const os = goos();
  const arch = goarch();
  const ext = os === "windows" ? ".zip" : ".tar.gz";
  const preferred = `clockify_${VERSION ? VERSION.replace(/^v/, "") + "_" : ""}${os}_${arch}${ext}`;
  // GoReleaser name template: clockify_{version}_{os}_{arch}
  const candidates = assets.filter((a) => {
    const n = a.name.toLowerCase();
    return n.includes(os) && n.includes(arch) && (n.endsWith(".tar.gz") || n.endsWith(".zip"));
  });
  if (!candidates.length) {
    throw new Error(
      `no asset for ${os}/${arch} in release (have: ${assets.map((a) => a.name).join(", ")})`
    );
  }
  const exact = candidates.find((a) => a.name === preferred);
  return exact || candidates[0];
}

async function extractBinary(archivePath, binPath) {
  const os = goos();
  if (archivePath.endsWith(".zip")) {
    // Use system unzip when available; otherwise fail with guidance.
    execSync(`unzip -o -j "${archivePath}" clockify* -d "${path.dirname(binPath)}"`, {
      stdio: "inherit",
    });
    // normalize name
    const extracted = path.join(path.dirname(binPath), os === "windows" ? "clockify.exe" : "clockify");
    if (extracted !== binPath) fs.renameSync(extracted, binPath);
  } else if (archivePath.endsWith(".tar.gz")) {
    execSync(`tar -xzf "${archivePath}" -C "${path.dirname(binPath)}"`, { stdio: "inherit" });
    const extracted = path.join(path.dirname(binPath), "clockify");
    if (extracted !== binPath) fs.renameSync(extracted, binPath);
  } else {
    fs.copyFileSync(archivePath, binPath);
  }
  fs.chmodSync(binPath, 0o755);
}

async function main() {
  const binDir = path.join(__dirname, "bin");
  fs.mkdirSync(binDir, { recursive: true });
  const binPath = path.join(binDir, "clockify");
  if (process.platform === "win32") {
    // npm bin shim uses .cmd; binary itself is clockify.exe
  }

  let release;
  if (VERSION) {
    const tag = VERSION.startsWith("v") ? VERSION : `v${VERSION}`;
    release = await getJSON(`${API_BASE}/repos/${REPO_OWNER}/${REPO_NAME}/releases/tags/${tag}`);
  } else {
    release = await getJSON(`${API_BASE}/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest`);
  }

  const asset = pickAsset(release.assets || []);
  const tmp = path.join(binDir, asset.name);
  await download(asset.browser_download_url, tmp);
  await extractBinary(tmp, binPath);
  try {
    fs.unlinkSync(tmp);
  } catch (_) {
    /* ignore */
  }
  console.log(`clockify-cli: installed ${release.tag_name} -> ${binPath}`);
}

main().catch((err) => {
  console.error("clockify-cli postinstall failed:", err.message);
  console.error("You can install manually from https://github.com/" + REPO_OWNER + "/" + REPO_NAME + "/releases");
  // Do not hard-fail npm install; users can still invoke via npx after manual install.
  process.exitCode = 0;
});
