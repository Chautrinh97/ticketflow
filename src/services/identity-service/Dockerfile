# Build context is src/ (see deployments/docker-compose.yaml) — every
# service needs the whole module (shared go.mod, src/pkg, src/proto).
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/identity-service ./services/identity-service/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/identity-service /usr/local/bin/identity-service
ENTRYPOINT ["/usr/local/bin/identity-service"]
