# Stage 1: Go builder
FROM golang:1.25-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o sana .

# Stage 2: Frontend builder
FROM node:24-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/bun.lock* ./
RUN npm install -g bun && bun install --frozen-lockfile
COPY frontend/ ./
RUN bun run build

# Stage 3: Runtime
FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /build/sana ./sana
COPY --from=frontend-builder /frontend/dist ./frontend/dist/
EXPOSE 8080
CMD ["./sana"]
