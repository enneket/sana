FROM alpine:3.19
WORKDIR /app
COPY backend/sana ./sana
COPY frontend/dist/ ./frontend/dist/
EXPOSE 8080
CMD ["./sana"]
