FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o sandboxapi-mcp .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/sandboxapi-mcp /usr/local/bin/sandboxapi-mcp

EXPOSE 8081
ENTRYPOINT ["sandboxapi-mcp"]
