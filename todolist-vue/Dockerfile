# syntax=docker/dockerfile:1

# -----------------------------
# Build stage: compile Vue app
# -----------------------------
FROM node:18-alpine AS build

WORKDIR /app

# Install dependencies first (better layer caching)
COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
# Prefer npm; adapt if lockfile exists
RUN \
    if [ -f package-lock.json ]; then npm ci; \
    elif [ -f yarn.lock ]; then yarn install --frozen-lockfile; \
    elif [ -f pnpm-lock.yaml ]; then npm i -g pnpm && pnpm i --frozen-lockfile; \
    else npm i; fi

# Copy source
COPY . .

# Build-time configurable API base URL for Vite
# Vite exposes VITE_* at build time; default to "/" (use dev proxy semantics)
ARG VITE_API_BASE_URL=/
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

# Build production assets
RUN npm run build

# --------------------------------
# Runtime stage: serve with Node.js
# --------------------------------
FROM node:18-alpine AS runtime

WORKDIR /app

# Labels
LABEL org.opencontainers.image.title="TodoList Vue Frontend" \
    org.opencontainers.image.description="Minimal Vue 3 app for TodoList API, served with Node.js" \
    org.opencontainers.image.source="todolist/todolist-vue" \
    org.opencontainers.image.licenses="MIT"

ENV NODE_ENV=production
ENV PORT=8080

# Install minimal runtime dependency for the server
RUN npm init -y && npm i express@4

# Copy built assets and server
COPY --from=build /app/dist ./dist
COPY --from=build /app/static ./static
COPY --from=build /app/server.js ./server.js

EXPOSE 8080

CMD ["node", "server.js"]
