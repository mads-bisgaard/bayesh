FROM bats/bats:latest AS base

RUN apk update; \
    apk add \
    python3 \
    sqlite \
    jq \
    inotify-tools \
    fzf \
    sudo