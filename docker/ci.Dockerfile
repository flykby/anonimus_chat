# syntax=docker/dockerfile:1

# CI runner image: lint, test, and docker build via host socket mount.
FROM golang:1.22-bookworm

ARG GOLANGCI_LINT_VERSION=v1.62.2
ARG DOCKER_VERSION=24.0.9

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    make \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b /usr/local/bin "${GOLANGCI_LINT_VERSION}"

RUN curl -fsSL "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" \
    | tar xz --strip-components=1 -C /usr/local/bin docker/docker

WORKDIR /opt/deps
COPY requirements.txt pyproject.toml ./
RUN pip3 install --break-system-packages --no-cache-dir -r requirements.txt

WORKDIR /workspace
ENV CI=true
ENV PATH="/usr/local/go/bin:${PATH}"

ENTRYPOINT ["bash"]
