# Estágio de compilação
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Estágio final (Execução)
FROM alpine:latest

# Instala dependências críticas: certificados SSL, ffmpeg e yt-dlp
RUN apk add --no-cache \
    ca-certificates \
    ffmpeg \
    python3 \
    py3-pip \
    && python3 -m pip install --no-cache-dir yt-dlp

# Verifica que as dependências foram instaladas corretamente
RUN which ffmpeg yt-dlp || (echo "ERRO: ffmpeg ou yt-dlp não encontrado" && exit 1)

WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static

# Cria a pasta de gravações com permissões apropriadas
RUN mkdir -p ./recordings && chmod 755 ./recordings

EXPOSE 8080

# Health check integrado (Google Cloud verifica esta URL)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]