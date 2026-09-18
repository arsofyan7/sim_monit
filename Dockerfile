# ==========================================
# STAGE 1: Frontend Build (Vue 3 + Vite)
# ==========================================
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# ==========================================
# STAGE 2: Backend Build (Go 1.22+)
# ==========================================
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app

# Install git/ca-certificates if needed
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compile static binary without CGO (using modernc.org/sqlite)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/sim_monit main.go

# ==========================================
# STAGE 3: Final Runner Container
# ==========================================
FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata openssh-client

# Copy Go binary from backend-builder
COPY --from=backend-builder /app/sim_monit /app/sim_monit

# Copy built frontend assets from frontend-builder
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Expose server port
EXPOSE 8080

ENV PORT=8080 \
    DB_PATH=/app/monitoring.db \
    GIN_MODE=release \
    STATIC_DIR=/app/frontend/dist

CMD ["/app/sim_monit"]
