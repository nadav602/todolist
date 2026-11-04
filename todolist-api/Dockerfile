# Multi-stage build for Go todos API
FROM golang:1.24 AS builder
WORKDIR /src
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/todos-api

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /out/todos-api /app/todos-api
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/app/todos-api"]
