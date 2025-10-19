# syntax=docker/dockerfile:1.7

FROM emscripten/emsdk:3.1.73 AS builder

ARG GO_VERSION=1.23.0
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
  build-essential \
  cmake \
  curl \
  git \
  python3 \
  && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
  | tar -C /usr/local -xz

ENV PATH=/usr/local/go/bin:${PATH}

WORKDIR /app
COPY . .

RUN make clean && make all

FROM scratch AS artifacts
COPY --from=builder /app/dist /dist
