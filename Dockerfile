# Stage 1: Build Vue frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web/app
COPY web/app/package*.json ./
RUN npm ci --prefer-offline
COPY web/app/ ./
RUN npm run build
# Output goes to web/static/ (set in vue.config.js outputDir: '../static')

# Stage 2: Build Go binary
FROM golang:alpine AS builder
RUN apk --update add ca-certificates
WORKDIR /app
COPY . ./
# Overwrite static dir with freshly built frontend
COPY --from=frontend /app/web/static ./web/static
RUN go mod tidy -diff
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gatus .

# Stage 3: Minimal runtime image (no config baked in — always mount at runtime)
FROM scratch
COPY --from=builder /app/gatus .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENV GATUS_CONFIG_PATH="/config/config.yaml"
ENV GATUS_LOG_LEVEL="INFO"
ENV PORT="8080"
EXPOSE 8080
ENTRYPOINT ["/gatus"]
