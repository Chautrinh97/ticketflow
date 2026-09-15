FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/payment-service ./services/payment-service/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/payment-service /usr/local/bin/payment-service
ENTRYPOINT ["/usr/local/bin/payment-service"]
