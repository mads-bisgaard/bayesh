FROM debian:stable-slim AS base

RUN apt-get update; \
    apt-get install -y \
    bats \
    curl \
    git \
    grep \
    inotify-tools \
    jq \
    make \
    sqlite3 \
    sudo \
    zsh

# Bayesh needs a newer fzf than the one packaged by apt. It goes in /usr/bin,
# since the bayesh binary is mounted over /usr/local/bin.
ARG FZF_VERSION=0.74.5
RUN asset="fzf-${FZF_VERSION}-linux_$(dpkg --print-architecture).tar.gz" \
    && url="https://github.com/junegunn/fzf/releases/download/v${FZF_VERSION}" \
    && cd /tmp \
    && curl -fsSLO "${url}/${asset}" \
    && curl -fsSL "${url}/fzf_${FZF_VERSION}_checksums.txt" | grep -F "${asset}" | sha256sum -c - \
    && tar -xzf "${asset}" -C /usr/bin fzf \
    && rm "${asset}"

RUN git clone https://github.com/ztombol/bats-support batslib/bats-support && \
    git clone https://github.com/ztombol/bats-assert batslib/bats-assert

ENV BATS_LIB_PATH=batslib
WORKDIR /code
ENTRYPOINT ["bats"]
CMD ["--help"]