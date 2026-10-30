FROM golang:1.24.3-bookworm AS go-toolchain

FROM ubuntu:24.04

COPY --from=go-toolchain /usr/local/go /usr/local/go

ENV PATH="/usr/local/go/bin:/usr/local/bin:/home/ubuntu/.local/bin:/home/ubuntu/.claude/bin:${PATH}" \
    HOME=/home/ubuntu \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

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
    locales \
    make \
    pre-commit \
    python3 \
    sqlite3 \
    sudo \
    tmux \
    zsh \
    && sed -i 's/^# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen \
    && locale-gen \
    && rm -rf /var/lib/apt/lists/*

RUN usermod --shell /usr/bin/zsh ubuntu \
    && echo "ubuntu ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/ubuntu \
    && chmod 0440 /etc/sudoers.d/ubuntu

RUN mkdir -p /batslib \
    && git clone --depth 1 https://github.com/ztombol/bats-support /batslib/bats-support \
    && git clone --depth 1 https://github.com/ztombol/bats-assert /batslib/bats-assert

ENV BATS_LIB_PATH=/batslib

USER ubuntu
RUN git clone --depth 1 https://github.com/ohmyzsh/ohmyzsh.git "${HOME}/.oh-my-zsh" \
    && cp "${HOME}/.oh-my-zsh/templates/zshrc.zsh-template" "${HOME}/.zshrc"

RUN curl -fsSL https://claude.ai/install.sh | bash

RUN git config --global --add safe.directory /workspaces/bayesh

WORKDIR /workspaces/bayesh
CMD ["/bin/zsh"]
