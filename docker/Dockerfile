# Stage 1: builder
FROM golang:1.25-alpine AS builder
ENV GOTOOLCHAIN=auto
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o sana .

# Stage 2: runtime
FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /build/sana ./sana
COPY frontend/dist/ ./frontend/dist/
EXPOSE 8080
CMD ["./sana"]
