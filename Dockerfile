# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o ai-mgr .

# Build UI stage
FROM --platform=$BUILDPLATFORM node:20-alpine AS ui-builder

WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci

COPY ui/ .
RUN npm run build

# Production stage
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk --no-cache add ca-certificates bash

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -u 1000 -G app -s /bin/bash -D app

# Copy binaries and files from builder stages
COPY --from=builder /app/ai-mgr /usr/local/bin/
COPY --from=builder /app/ui/dist /opt/ai-manager/web-ui
COPY --from=builder /app/README.md /opt/ai-manager/
COPY --from=builder /app/LICENSE /opt/ai-manager/

# Create config directory
RUN mkdir -p /home/app/.ai-manager && \
    chown -R app:app /home/app/.ai-manager /opt/ai-manager

# Switch to non-root user
USER app

# Expose ports
EXPOSE 19999

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:19999/api/health || exit 1

# Run the daemon
ENTRYPOINT ["ai-mgr", "daemon"]
