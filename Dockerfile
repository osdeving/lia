# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instalar dependências de build necessárias
RUN apk add --no-cache git

# Copiar go.mod e go.sum para cache de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar o resto do código fonte
COPY . .

# Build do binário 'lia'
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/lia ./cmd/lia/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copiar o binário do builder
COPY --from=builder /app/lia /usr/local/bin/lia

# Copiar diretórios de specs e exemplos para o container
COPY --from=builder /app/docs/spec /etc/lia/spec
COPY --from=builder /app/examples /app/examples

# Variáveis de ambiente padrão
ENV LIA_SPEC_PATH=/etc/lia/spec

# Comando padrão
ENTRYPOINT ["lia"]
CMD ["--help"]
