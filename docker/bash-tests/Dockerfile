FROM debian:stable-slim AS base

RUN apt-get update; \
    apt-get install -y \
    git \
    bats \
    python3 \
    sqlite3 \
    jq \
    inotify-tools \
    fzf \
    sudo \
    grep \
    curl \
    jq

RUN git clone https://github.com/ztombol/bats-support batslib/bats-support && \
    git clone https://github.com/ztombol/bats-assert batslib/bats-assert

ENV BATS_LIB_PATH=batslib
WORKDIR /code
ENTRYPOINT ["bats"]
CMD ["--help"]