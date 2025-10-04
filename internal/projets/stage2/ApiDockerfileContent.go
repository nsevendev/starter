package stage2

import "fmt"

func ApiDockerfileContent() string {
	return fmt.Sprintf(`FROM golang:1.24.4-bookworm AS base
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*
RUN go install github.com/air-verse/air@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go install github.com/nsevenpack/mignosql/cmd/migrationcreate@latest
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
RUN mkdir -p /app/tmp/air
RUN mkdir -p /app/tmp/air/api
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
ARG SERVICE=api
ENV SERVICE=${SERVICE}

FROM base AS dev
WORKDIR /app
COPY . .
CMD ["sh", "-c", "air -c .air.toml"]

FROM base AS build
WORKDIR /app
COPY . .
RUN swag init -o docs -g cmd/${SERVICE}/main.go --parseInternal --pd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/${SERVICE} ./cmd/${SERVICE}
CMD ["sh", "-c", "ls -l /app/dist/${SERVICE}"]

FROM golang:1.24.4-bookworm AS runtime-base
WORKDIR /app
# certificats (pour les appels HTTPS éventuels)
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
ARG SERVICE=api
ENV SERVICE=${SERVICE}
COPY --from=build /app/dist/${SERVICE} /app/application
COPY go.mod .
RUN chmod +x /app/application
CMD ["./application"]

FROM runtime-base AS prod
FROM runtime-base AS preprod
`)
}
