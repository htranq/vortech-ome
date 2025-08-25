FROM golang:1.24-alpine AS builder
WORKDIR /workspace

# Install certs for module fetching if needed
RUN apk add --no-cache ca-certificates git

# Cache deps first
COPY go.mod go.sum ./
COPY vendor ./vendor
RUN --mount=type=cache,target=/go/pkg/mod true

# Copy the rest of the source
COPY . .

# Build static binary for the target platform
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -mod=vendor -o /out/server ./cmd/main.go

# Final stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /out/server /app/server
COPY configs/config.yaml /app/config.yaml

EXPOSE 8080 8081

CMD ["/app/server", "-c", "/app/config.yaml"]
