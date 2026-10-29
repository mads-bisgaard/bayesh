FROM golang:1.24.3-bookworm AS go-toolchain
FROM node:22-bookworm AS node-toolchain

FROM ubuntu:24.04

ARG LOGSEQ_CLI_VERSION=0.4.3

COPY --from=go-toolchain /usr/local/go /usr/local/go
COPY --from=node-toolchain /usr/local/ /usr/local/

ENV PATH="/usr/local/go/bin:/usr/local/bin:/home/work/.local/bin:/home/work/.claude/bin:${PATH}" \
    HOME=/home/work \
    LOGSEQ_GRAPH=bayesh-wiki

RUN apt-get update && apt-get install -y --no-install-recommends \
    bats \
    build-essential \
    ca-certificates \
    curl \
    fzf \
    git \
    inotify-tools \
    jq \
    libatomic1 \
    libsqlite3-dev \
    make \
    python3 \
    sqlite3 \
    sudo \
    tmux \
    zsh \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd work \
    && useradd --gid work --create-home --shell /usr/bin/zsh work \
    && echo "work ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/work \
    && chmod 0440 /etc/sudoers.d/work

RUN npm install --global "@logseq/cli@${LOGSEQ_CLI_VERSION}"

RUN mkdir -p /batslib \
    && git clone --depth 1 https://github.com/ztombol/bats-support /batslib/bats-support \
    && git clone --depth 1 https://github.com/ztombol/bats-assert /batslib/bats-assert

ENV BATS_LIB_PATH=/batslib

RUN mkdir -p /home/work/logseq \
    && chown -R work:work /home/work

USER work
RUN curl -fsSL https://claude.ai/install.sh | bash

WORKDIR /workspaces/bayesh
CMD ["/bin/bash"]
