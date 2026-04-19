FROM golang:1.26-alpine AS deps
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

FROM deps AS builder
COPY . .
RUN go generate ./... \
    && CGO_ENABLED=0 GOOS=linux go build \
        -trimpath -ldflags="-s -w" \
        -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
    && adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=builder /out/api /app/api
EXPOSE 8080
USER app
ENTRYPOINT ["/app/api"]
