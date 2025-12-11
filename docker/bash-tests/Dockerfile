FROM debian:stable-slim AS base

RUN apt-get update; \
    apt-get install -y \
    bats \
    curl \
    fzf \
    git \
    grep \
    inotify-tools \
    jq \
    make \
    sqlite3 \
    sudo \
    zsh

RUN git clone https://github.com/ztombol/bats-support batslib/bats-support && \
    git clone https://github.com/ztombol/bats-assert batslib/bats-assert

ENV BATS_LIB_PATH=batslib
WORKDIR /code
ENTRYPOINT ["bats"]
CMD ["--help"]