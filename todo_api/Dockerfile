# stage 1 — build
# берем образ, даем через AS название, чтобы далее обратиться
FROM golang:1.24-alpine AS builder 

# папка для билда проекта на втором стэйдже
WORKDIR /build

# кешируем зависимости, чтобы не качать их каждый раз при изменении кода
COPY go.mod go.sum ./
RUN go mod download

# копируем весь код и билдим бинарник
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /build/server ./cmd/main.go

# stage 2 — runtime
FROM gcr.io/distroless/base-debian12

# копируем бинарник из первого стэйджа в диреккторию /app
WORKDIR /app
COPY --from=builder /build/server /app/server

# создаем непривилегированного пользователя
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
