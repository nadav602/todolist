/**
 * Express-based Node.js server for serving the Vue TodoList app.
 *
 * Features:
 * - Serves the entire SPA at the root path
 * - Path redirection should be handled by the ingress (e.g., Traefik)
 * - Runtime config.js exposes API_BASE_URL and TITLE via window.APP_CONFIG
 * - Simple express.static mounting handles asset serving
 *
 * Environment variables:
 * - API_BASE_URL: REQUIRED; Full URL or path for API calls (e.g., "/todos" or "http://api.example.com")
 * - HTML_DIR: Built dist directory (default: ./dist or /usr/share/nginx/html)
 * - STATIC_DIR: Static assets directory (default: ./static or {HTML_DIR}/static)
 * - USER: Optional; sets window.APP_CONFIG.TITLE
 * - PORT: Server port (default: 8080)
 */

const fs = require("fs");
const path = require("path");
const express = require("express");

const app = express();

// ============================================================================
// Configuration
// ============================================================================

const PORT = parseInt(process.env.PORT || "8080", 10);
const API_BASE_URL = process.env.API_BASE_URL;
const USERNAME = process.env.USER || "";
const TITLE = USERNAME ? `${USERNAME}'s TODOS` : "My TODOS";

console.log(
  `[startup] Looking for title image at: /opt/images/title-image.png`,
);

if (!API_BASE_URL) {
  console.error(
    "ERROR: API_BASE_URL environment variable is required but not set",
  );
  process.exit(1);
}

// Resolve directories
const LOCAL_DIST_DIR = path.join(__dirname, "dist");
const FALLBACK_BUILD_DIR = "/usr/share/nginx/html";
const DEFAULT_HTML_DIR = fs.existsSync(LOCAL_DIST_DIR)
  ? LOCAL_DIST_DIR
  : FALLBACK_BUILD_DIR;
const HTML_DIR = process.env.HTML_DIR || DEFAULT_HTML_DIR;

const DEFAULT_STATIC_DIR = fs.existsSync(path.join(__dirname, "static"))
  ? path.join(__dirname, "static")
  : path.join(HTML_DIR, "static");
const STATIC_DIR = process.env.STATIC_DIR || DEFAULT_STATIC_DIR;

const INDEX_HTML_PATH = path.join(HTML_DIR, "index.html");

// ============================================================================
// Runtime config generation
// ============================================================================

function generateRuntimeConfig() {
  const apiBase = JSON.stringify(API_BASE_URL);
  const title = JSON.stringify(TITLE);
  return `(function() {
  window.APP_CONFIG = window.APP_CONFIG || {};
  window.APP_CONFIG.API_BASE_URL = ${apiBase};
  window.APP_CONFIG.TITLE = ${title};
})();\n`;
}

const runtimeConfig = generateRuntimeConfig();

// ============================================================================
// Load and cache index.html
// ============================================================================

let indexHtmlContent = null;

if (fs.existsSync(INDEX_HTML_PATH)) {
  indexHtmlContent = fs.readFileSync(INDEX_HTML_PATH, "utf8");
} else {
  // Minimal fallback
  indexHtmlContent = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>${TITLE}</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <script src="config.js"></script>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="src/main.js"></script>
</body>
</html>`;
}

// ============================================================================
// Middleware
// ============================================================================

app.use((req, res, next) => {
  const start = Date.now();
  res.on("finish", () => {
    const ms = Date.now() - start;
    console.log(`${req.method} ${req.originalUrl} ${res.statusCode} ${ms}ms`);
  });
  next();
});

// ============================================================================
// Routes
// ============================================================================

// Serve static files (assets, images, etc.)
app.use(express.static(HTML_DIR, { index: false }));

// Serve the /static directory (project static files)
app.use("/static", express.static(STATIC_DIR, { index: false }));

// Serve runtime config.js
app.get("/config.js", (req, res) => {
  res.type("application/javascript; charset=utf-8");
  res.set(
    "Cache-Control",
    "no-store, no-cache, must-revalidate, proxy-revalidate",
  );
  res.send(runtimeConfig);
});

// Serve favicon
app.get("/favicon.ico", (req, res) => {
  const faviconPath = path.join(STATIC_DIR, "img", "todolist.ico");
  res.sendFile(faviconPath, (err) => {
    if (err) res.status(404).end();
  });
});

// Serve title image from file if it exists
app.get("/title-image.png", (req, res) => {
  const titleImagePath = "/opt/images/title-image.png";
  console.log(`[title-image.png] Attempting to read: ${titleImagePath}`);
  res.sendFile(titleImagePath, (err) => {
    if (err) {
      console.error(
        `[title-image.png] Error reading file: ${err.code} - ${err.message}`,
      );
      console.log(`[title-image.png] Falling back to todolist.ico`);
      const fallbackPath = path.join(__dirname, "static/img/todolist.ico");
      res.sendFile(fallbackPath, (fallbackErr) => {
        if (fallbackErr) {
          console.error(
            `[title-image.png] Error reading fallback file: ${fallbackErr.code} - ${fallbackErr.message}`,
          );
          return res.status(404).json({ error: "Title image not found" });
        }
      });
    }
  });
});

// SPA fallback: serve index.html for all other requests
app.get("*", (req, res) => {
  // Don't serve index.html for direct file requests (has extension)
  if (/\.[a-zA-Z0-9]+$/.test(req.path)) {
    return res.status(404).end();
  }
  res.type("text/html; charset=utf-8").send(indexHtmlContent);
});

// ============================================================================
// Start server
// ============================================================================

app.listen(PORT, "0.0.0.0", () => {
  console.log("TodoList Vue server is running");
  console.log(`- PORT: ${PORT}`);
  console.log(`- HTML_DIR: ${path.resolve(HTML_DIR)}`);
  console.log(`- STATIC_DIR: ${path.resolve(STATIC_DIR)}`);
  console.log(`- API_BASE_URL: ${API_BASE_URL}`);
  console.log(`- TITLE: ${TITLE}`);
});
