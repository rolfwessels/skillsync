# ── base: shared Go toolchain ─────────────────────────────────────────────────
FROM golang:1.26-alpine AS base

RUN apk update \
  && apk upgrade \
  && apk add --no-cache \
    ca-certificates \
    git \
    curl \
    wget \
    bash \
    make \
    rsync \
    nano \
    zsh \
    zsh-vcs \
    docker-cli \
    docker-cli-compose \
    openssh \
  && update-ca-certificates \
  && git config --global --add safe.directory /skillsync

# ── build: compile the binary ─────────────────────────────────────────────────
FROM base AS build

ARG VERSION=dev
WORKDIR /skillsync

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X 'github.com/rolfwessels/skillsync/internal/cli.Version=${VERSION}'" \
    -o /out/skillsync ./cmd/skillsync

# ── runtime: minimal production image ─────────────────────────────────────────
FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates git
COPY --from=build /out/skillsync /usr/local/bin/skillsync
ENTRYPOINT ["skillsync"]

# ── dev: full dev environment (used by docker-compose) ────────────────────────
FROM base AS dev

ARG UID=1000
ARG GID=1000

RUN addgroup -g ${GID} dev && \
    adduser -D -u ${UID} -G dev -s /bin/zsh dev && \
    mkdir -p /skillsync && chown dev:dev /skillsync

ENV HOME=/home/dev
ENV GOPATH=/home/dev/go
ENV GOCACHE=/home/dev/.cache/go-build
ENV PATH="/home/dev/go/bin:${PATH}"
ENV TERM=xterm-256color

# oh-my-zsh (run as root targeting HOME=/home/dev, then fix ownership)
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended && \
    git clone --depth=1 https://github.com/zsh-users/zsh-autosuggestions \
      /home/dev/.oh-my-zsh/custom/plugins/zsh-autosuggestions && \
    chown -R dev:dev /home/dev

USER dev

# goreleaser (for snapshot builds)
RUN go install github.com/goreleaser/goreleaser/v2@latest

# golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

WORKDIR /skillsync
