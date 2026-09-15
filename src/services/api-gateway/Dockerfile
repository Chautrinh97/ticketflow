FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api-gateway ./services/api-gateway/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/api-gateway /usr/local/bin/api-gateway
ENTRYPOINT ["/usr/local/bin/api-gateway"]
