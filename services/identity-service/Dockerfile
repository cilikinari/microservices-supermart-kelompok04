# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY pkg/ ./pkg/
COPY services/identity-service/ ./services/identity-service/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /bin/identity-service ./services/identity-service/cmd

RUN adduser -D -g '' -u 10001 appuser

# STAGE 2: Minimal Runtime Environment (Distroless Non-Root)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/identity-service /app/identity-service

USER nonroot:nonroot

EXPOSE 3001

ENTRYPOINT ["/app/identity-service"]
