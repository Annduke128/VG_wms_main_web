# ============================================
# WMS v1 — API Server + Frontend
# Multi-stage build
# ============================================

# ---- Stage 1: Build Frontend ----
FROM node:20-alpine AS frontend

WORKDIR /web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ .
RUN npm run build

# ---- Stage 2: Build Go binary ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /api ./cmd/api

# ---- Stage 3: Runtime ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Non-root user
RUN addgroup -S wms && adduser -S wms -G wms

WORKDIR /app

COPY --from=builder /api .
COPY --from=frontend /web/dist web/dist/
COPY migrations/ migrations/

RUN mkdir -p /app/uploads
RUN chown -R wms:wms /app
USER wms

ENV STATIC_DIR=web/dist

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./api"]
