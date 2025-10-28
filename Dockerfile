FROM  golang:1.25.3-alpine3.22  AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o upload-notify ./cmd/server

FROM scratch

COPY --from=builder /app/upload-notify /upload-notify

COPY  --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs
EXPOSE 8080

CMD ["./upload-notify"]