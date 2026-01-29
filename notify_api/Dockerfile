FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY notify_api/go.mod notify_api/go.sum ./
COPY pkg ../pkg/
RUN go mod download

COPY notify_api/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /build/server ./cmd/main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /build/server /app/server

USER nonroot:nonroot
ENTRYPOINT ["/app/server"]