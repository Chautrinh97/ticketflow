FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/event-service ./services/event-service/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/event-service /usr/local/bin/event-service
ENTRYPOINT ["/usr/local/bin/event-service"]
