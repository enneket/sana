FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY ../backend/ ./
RUN go build -o sana .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/sana ./
COPY ../notes/ ./notes/
RUN apk add --no-cache sqlite-libs
EXPOSE 8080
CMD ["./sana"]
