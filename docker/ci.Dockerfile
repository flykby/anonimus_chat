# syntax=docker/dockerfile:1

FROM golang:1.22-bookworm

ARG GOLANGCI_LINT_VERSION=v1.62.2
ARG DOCKER_VERSION=24.0.9

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl git make && rm -rf /var/lib/apt/lists/*

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b /usr/local/bin "${GOLANGCI_LINT_VERSION}"

RUN curl -fsSL "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" \
    | tar xz --strip-components=1 -C /usr/local/bin docker/docker

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

WORKDIR /workspace
ENV CI=true
ENV PATH="/go/bin:/usr/local/go/bin:${PATH}"

ENTRYPOINT ["bash"]
