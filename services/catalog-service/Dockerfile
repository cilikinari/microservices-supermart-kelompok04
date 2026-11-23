# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY services/catalog-service/ ./services/catalog-service/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /bin/catalog-service ./services/catalog-service/cmd

RUN adduser -D -g '' -u 10001 appuser

# STAGE 2: Minimal Runtime Environment (Distroless Non-Root)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/catalog-service /app/catalog-service

USER nonroot:nonroot

EXPOSE 3002 50052

ENTRYPOINT ["/app/catalog-service"]
