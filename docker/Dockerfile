FROM debian:stable-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && apt-get clean
COPY backend/sana ./sana
COPY frontend/dist/ ./frontend/dist/
EXPOSE 8080
CMD ["./sana"]
