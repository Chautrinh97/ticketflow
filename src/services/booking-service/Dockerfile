FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/booking-service ./services/booking-service/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/booking-service /usr/local/bin/booking-service
ENTRYPOINT ["/usr/local/bin/booking-service"]
