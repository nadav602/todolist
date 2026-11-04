# API Service for Todo List demo App

## Go REST API service (api)

This repo also includes a lightweight Go service that provides RESTful APIs for todolist application 

Endpoints:
- `GET /todos` (list), 
- `GET /todos/{id}` (fetch one)
- `POST /todos` (create; body: `{ "task": "..." }`, status defaults to `"Todo"`, id generated as `MAX(id)+1` within a transaction)
- `PUT/PATCH /todos/{id}` (update `task` and/or `status`; valid statuses: `"Todo"`, `"In Progress"`, `"Complete"`; response shows `updated_fields` or `"no changes"`)
- `DELETE /todos/{id}` (remove)
- `GET /livez`, `GET /readyz` health probes

## API Usage Examples with curl, assuming service is running locally 

### List all todos

```bash
curl http://localhost:8080/todos
```

### Get a specific todo

```bash
curl http://localhost:8080/todos/1
```

### Create a new todo

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"task": "Buy groceries"}'
```

### Update a todo (PUT)

```bash
# Update both task and status
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"task": "Buy groceries and cook", "status": "In Progress"}'
```

### Update a todo (PATCH)

```bash
# Update only status
curl -X PATCH http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"status": "Complete"}'

# Update only task
curl -X PATCH http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"task": "Buy organic groceries"}'
```

### Delete a todo

```bash
curl -X DELETE http://localhost:8080/todos/1
```

## Running the API

Run locally:

```bash
cd api
export MYSQL_USER=<user>
export MYSQL_PASSWORD=<password>
export MYSQL_DB=<database>
export MYSQL_HOST=<host-or-host:port>
export PORT=8080  # optional
go run .
```

Initialize the configured todo table during startup only when needed:

```bash
go run . --init-schema
```

Build and run the Go API as a container:

```bash
cd api
docker build -t todos-go-api .
docker run --rm -p 8080:8080 \
  -e MYSQL_USER=<user> \
  -e MYSQL_PASSWORD=<password> \
  -e MYSQL_DB=<database> \
  -e MYSQL_HOST=<host-or-host:port> \
  todos-go-api
```

## Building and Pushing Container Images

A `Makefile` is provided to simplify building and pushing container images.

### Available Targets

#### Build Targets
- `make build` - Build REST API container image  

#### Combined Build & Push Targets
- `make build-push` - Build and push API

### Usage

#### Using Environment Variable (Recommended)

Set the container registry once:

```bash
export CONTAINER_REGISTRY=ghcr.io/bennyro-mta
```

Then build and push:

```bash
# Build and push
make build-push VERSION=<##.###>
```

When prompted for registry prefix, press Enter to use the default from `CONTAINER_REGISTRY`.

#### Using Command-Line Arguments

```bash
# Build only
make build VERSION=1.10 PREFIX=ghcr.io/bennyro-mta

# Build and push in one command
make build-push VERSION=1.10 PREFIX=docker.io/myuser
```

### Parameters

- **VERSION**: Container image version/tag (required for build/push)
  - Prompted if not provided: `Enter version for images:`
  - Example: `1.10`, `1.0.0`, `latest`

- **PREFIX**: Container registry and project path (optional for push)
  - If not provided on CLI, uses `CONTAINER_REGISTRY` env var if set
  - If env var not set, prompts interactively: `Enter registry prefix [default]:`
  - Example: `ghcr.io/bennyro-mta`, `docker.io/myuser`, `localhost:5000`
# test
# trivy test
