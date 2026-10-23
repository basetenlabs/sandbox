ARG SANDBOX_VERSION=latest
FROM docker.io/baseten/sandbox:${SANDBOX_VERSION} AS sandbox-api

FROM node:24-slim@sha256:55fb5f5a527ea2c5c4a95f963e48ef90aa38c7b346054fdc8ee9957a2b55d205

RUN apt-get update && apt-get install -y --no-install-recommends \
  bash \
  ca-certificates \
  curl \
  git \
  ripgrep \
  && rm -rf /var/lib/apt/lists/*

RUN npm install -g @openai/codex@0.140.0

WORKDIR /workspace

COPY --from=sandbox-api /sandbox-api /usr/local/bin/sandbox-api
RUN chmod +x /usr/local/bin/sandbox-api

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/sandbox-api"]
