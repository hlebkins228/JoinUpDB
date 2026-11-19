FROM golang:1.25.5-bookworm AS runner
ARG TARGETARCH
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
    && go build -o /app/migrate ./cmd/migrations \
    && CGO_ENABLED=0 GOOS=windows GOARCH=${TARGETARCH:-amd64} \
       go test -c -tags=e2e -o /app/e2e.test.exe ./internal/e2e_test
ENV GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
ENTRYPOINT ["bash", "tests/run.sh"]

FROM debian:bookworm-slim AS app
WORKDIR /app
COPY --from=runner /app/server /app/migrate ./
COPY tests/server.yaml ./
CMD ["sh", "-ec", "sed \"s/TEST_JWT_SECRET/$(cat /proc/sys/kernel/random/uuid)/\" server.yaml > config.yaml; ./migrate up; exec ./server"]
