# syntax=docker/dockerfile:1

FROM node:20-alpine AS assets
WORKDIR /build
COPY package.json ./
RUN npm install
COPY tailwind.config.js ./
COPY web/src ./web/src
COPY web/static/js ./web/static/js
COPY internal/templates ./internal/templates
COPY internal/api ./internal/api
COPY cmd ./cmd
RUN npm run build:css
RUN npm run build:js

FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=assets /build/web/static/css/tailwind.css ./web/static/css/tailwind.css
COPY --from=assets /build/web/static/js/htmx.min.js ./web/static/js/htmx.min.js
COPY --from=assets /build/web/static/js/alpine.min.js ./web/static/js/alpine.min.js
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o /out/lume ./cmd/lume

FROM alpine:latest AS runtime
WORKDIR /app
RUN apk add --no-cache tzdata ca-certificates
COPY --from=builder /out/lume /app/lume
COPY --from=builder /src/internal/templates /app/internal/templates
COPY --from=builder /src/internal/i18n /app/internal/i18n
COPY --from=builder /src/web/static /app/web/static
RUN mkdir -p /app/data
EXPOSE 8080
ENV DB_PATH=/app/data/lume.db
CMD ["/app/lume"]
