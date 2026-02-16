FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o proteus ./cmd/proteus

FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/proteus /app/proteus

EXPOSE 8080

ENTRYPOINT ["/app/proteus"]
