# TodoList Vue Frontend

A minimal Vue 3 + Vite frontend for the TodoList REST API.

## Prerequisites

- Node.js 18+ and npm
- The TodoList API running (default: http://localhost:8080)

## Quick Start

### Development

1. Install dependencies:
   ```bash
   npm install
   ```

2. Start the dev server:
   ```bash
   npm run dev
   ```
   - Opens at http://localhost:5173
   - Dev server proxies `/todos`, `/livez`, `/readyz` to http://localhost:8080 (no CORS issues)

### Production Build

1. Build for production:
   ```bash
   npm run build
   ```
   - Output in `dist/`

2. Preview the production build locally:
   ```bash
   npm run preview
   ```

## Running in Container

### Build the Docker Image

```bash
docker build -t todolist-vue:latest .
```

### Run the Container

```bash
docker run -p 8080:8080 \
  -e API_BASE_URL="/todos" \
  -e USER="benny" \
  todolist-vue:latest
```

Access at http://localhost:8080

#### With Custom Title Image

```bash
docker run -p 8080:8080 \
  -e API_BASE_URL="/todos" \
  -e USER="benny" \
  -e TITLE_IMAGE_BASE64="$(base64 -w 0 icon.png)" \
  todolist-vue:latest
```

#### Environment Variables

- `API_BASE_URL` (required): Full URL or path for API calls
  - Example: `/todos` (same host, proxied by reverse proxy)
  - Example: `http://localhost:8080` (direct URL)
- `USER` (optional): Sets the page title to "{USER}'s TODOS"
- `TITLE_IMAGE_BASE64` (optional): Base64-encoded image to display next to the title (64×64px)
  - Use a small image (the image is displayed at 64×64px regardless of original size)
  - Example: `TITLE_IMAGE_BASE64="$(base64 -w 0 icon.png)"`
- `PORT` (optional): Server port, defaults to `8080`

## API Configuration

The frontend calls the API at the path/URL specified in `API_BASE_URL`.

### Development

Dev proxy in `vite.config.js` handles CORS automatically. No special configuration needed if API is at http://localhost:8080.

### Production

Set `API_BASE_URL` to the API endpoint:

- **Same host, different path**: `API_BASE_URL=/todos`
- **Different host**: `API_BASE_URL=http://api.example.com`
- **Kubernetes service**: `API_BASE_URL=http://todo-api:8080`

## API Endpoints

The TodoList API provides:

- `GET /todos` — List all todos
- `GET /todos/{id}` — Get a specific todo
- `POST /todos` — Create a todo (body: `{ "task": "..." }`)
- `PATCH /todos/{id}` — Update a todo (body: `{ "task"?: string, "status"?: string }`)
- `DELETE /todos/{id}` — Delete a todo
- `GET /livez` — Liveness probe
- `GET /readyz` — Readiness probe

Supported statuses: `Todo`, `In Progress`, `Complete`

## Scripts

- `npm run dev` — Start Vite dev server with proxy
- `npm run build` — Build for production
- `npm run preview` — Preview production build

## Troubleshooting

**"Failed to load todos" error**
- Confirm API is running at the configured `API_BASE_URL`
- Check browser console for network errors
- In dev: Ensure `npm run dev` started successfully

**API not reachable in container**
- Verify `API_BASE_URL` is correct for your network setup
- If using Docker networking, use service name: `http://todo-api:8080`
- If behind a reverse proxy, set `API_BASE_URL` to the proxied path or full URL

## Title Image

The title image (64×64px) is displayed next to the page title in the header. By default, it shows the TodoList icon (`/favicon.ico`).

To use a custom image, pass it as a base64-encoded string:

```bash
docker run -p 8080:8080 \
  -e API_BASE_URL="/todos" \
  -e TITLE_IMAGE_BASE64="$(base64 -w 0 icon.png)" \
  todolist-vue:latest
```

**Supported Image Formats:** PNG, JPG, GIF, SVG, WebP, and any format browsers support.

While images are automatically scaled to 64×64px regardless of original size, it is recommended to provide a low-resolution image to bypass command line length limitations.

## Project Structure

- `index.html` — HTML entry point
- `src/main.js` — Vue app entry point
- `src/App.vue` — Main component (todos management)
- `vite.config.js` — Build and dev server config
- `server.js` — Express server for production
- `dist/` — Production build output
