# Build aşaması
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Go modüllerini kopyala ve indir
COPY go.mod go.sum ./
RUN go mod download

# Kaynak kodları kopyala
COPY . .

# Uygulamayı derle
RUN CGO_ENABLED=0 GOOS=linux go build -o casino-wallet-service ./cmd/api

# Çalışma aşaması
FROM alpine:latest

WORKDIR /app

# Sadece derlenmiş uygulamayı kopyala
COPY --from=builder /app/casino-wallet-service .

# Uygulamayı çalıştır
EXPOSE 8080
CMD ["./casino-wallet-service"] 