# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend-build
WORKDIR /src/frontend
COPY backend/internal/decoderlibrary/*.json /src/backend/internal/decoderlibrary/
COPY backend/internal/application/function_symbol_ecos_codes.json /src/backend/internal/application/
COPY tools/api-types/package.json tools/api-types/package-lock.json /src/tools/api-types/
COPY tools/frontend-lint/package.json tools/frontend-lint/package-lock.json /src/tools/frontend-lint/
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend ./
COPY openapi /src/openapi
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine AS backend-build
ARG TARGETOS
ARG TARGETARCH
ARG BUILD_COMMIT=unknown
ARG BUILD_KIND=development
ARG BUILD_MODIFIED=unknown
WORKDIR /src/backend
COPY backend/go.mod ./
RUN go mod download
COPY backend ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w -X railkeeper/backend/internal/buildinfo.commit=$BUILD_COMMIT -X railkeeper/backend/internal/buildinfo.kind=$BUILD_KIND -X railkeeper/backend/internal/buildinfo.modified=$BUILD_MODIFIED" -o /out/railkeeper ./cmd/railkeeper

FROM alpine:3.24 AS runtime
LABEL org.opencontainers.image.source="https://github.com/ichwars/RailKeeper" \
  org.opencontainers.image.licenses="AGPL-3.0-only"
RUN apk add --no-cache --upgrade \
  'libcrypto3>=3.5.8-r0' \
  'libssl3>=3.5.8-r0' \
  ca-certificates \
  tzdata \
  && adduser -D -H -u 10001 railkeeper \
  && mkdir -p /app/web /data \
  && chown -R railkeeper:railkeeper /app /data
COPY --from=backend-build /out/railkeeper /usr/local/bin/railkeeper
COPY --from=frontend-build /src/frontend/dist /app/web
COPY --chown=railkeeper:railkeeper backend/migrations /app/migrations
COPY --chown=railkeeper:railkeeper backend/seeds /app/seeds
RUN chmod -R a+rX /app
USER railkeeper
ENV RAILKEEPER_ADDR=:8080 \
  RAILKEEPER_DATA_DIR=/data \
  RAILKEEPER_MIGRATIONS_DIR=/app/migrations \
  RAILKEEPER_SEEDS_DIR=/app/seeds \
  RAILKEEPER_STATIC_DIR=/app/web
VOLUME ["/data"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD ["railkeeper", "healthcheck"]
ENTRYPOINT ["railkeeper"]
