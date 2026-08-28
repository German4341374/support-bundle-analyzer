# syntax=docker/dockerfile:1.10
FROM golang:1.26.7-alpine3.23 AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY apps/cli apps/cli
COPY apps/static-report-viewer apps/static-report-viewer
COPY internal internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/support-bundle-analyzer ./apps/cli

FROM node:24.14.1-alpine3.23 AS node-build
WORKDIR /src
COPY package.json package-lock.json ./
COPY apps/api/package.json apps/api/package.json
RUN npm ci --ignore-scripts
COPY apps/api apps/api
RUN npm run build && npm prune --omit=dev --ignore-scripts

FROM node:24.14.1-alpine3.23 AS runtime
ENV NODE_ENV=production \
    SBA_HOST=0.0.0.0 \
    SBA_PORT=8080 \
    SBA_ALLOW_REMOTE=true \
    SBA_RATE_LIMIT_MAX=120 \
    SBA_EXPENSIVE_RATE_LIMIT_MAX=10 \
    SBA_CORE_BINARY=/usr/local/bin/support-bundle-analyzer \
    SBA_INPUT_ROOT=/input \
    SBA_WORKSPACE_ROOT=/data/workspaces
RUN apk upgrade --no-cache \
    && rm -rf /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/corepack /opt/yarn-v1.22.22 \
    && rm -f /usr/local/bin/npm /usr/local/bin/npx /usr/local/bin/corepack /usr/local/bin/yarn /usr/local/bin/yarnpkg \
    && addgroup -S -g 10001 analyzer && adduser -S -D -H -u 10001 -G analyzer analyzer \
    && mkdir -p /app /data/workspaces /input \
    && chown -R analyzer:analyzer /app /data /input
WORKDIR /app
COPY --from=go-build --chown=root:root /out/support-bundle-analyzer /usr/local/bin/support-bundle-analyzer
COPY --from=node-build --chown=analyzer:analyzer /src/package.json /src/package-lock.json ./
COPY --from=node-build --chown=analyzer:analyzer /src/node_modules ./node_modules
COPY --from=node-build --chown=analyzer:analyzer /src/apps/api/dist ./apps/api/dist
USER 10001:10001
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD ["node", "-e", "fetch('http://127.0.0.1:8080/health').then(r=>{if(!r.ok)process.exit(1)}).catch(()=>process.exit(1))"]
CMD ["node", "apps/api/dist/main.js"]
