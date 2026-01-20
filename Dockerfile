# -----------------------
# Stage 1 — Build Go app
# -----------------------
FROM --platform=linux/amd64 golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/api

# -----------------------
# Stage 2 — Runtime image
# -----------------------
FROM --platform=linux/amd64 debian:bookworm-slim

# Install ONLY runtime OCR dependencies
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    tesseract-ocr-eng \
    libtesseract-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/app /app/app

EXPOSE 8080
CMD ["./app"]
