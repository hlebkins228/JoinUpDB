FROM golang:1.25.5-bookworm AS runner
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl jq tcpdump postgresql-client default-jre-headless nodejs npm \
    && rm -rf /var/lib/apt/lists/*
RUN npm install --prefix /opt/allure --no-package-lock --no-audit --no-fund allure-commandline@2.46.1
ENV PATH="/opt/allure/node_modules/.bin:${PATH}"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go \
    -d ./cmd/server,./internal/controller/handlers,./internal/controller/dto,./internal/models \
    -o docs --parseGoList=false \
    && go build -o /app/server ./cmd/server \
    && go build -o /app/migrate ./cmd/migrations
ENTRYPOINT ["bash", "tests/run.sh"]

FROM debian:bookworm-slim AS app
WORKDIR /app
COPY --from=runner /app/server /app/migrate ./
COPY tests/server.yaml tests/start-server.sh ./
ENTRYPOINT ["bash", "start-server.sh"]
