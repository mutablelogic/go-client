ARG OS
ARG ARCH

# Run makefile to build all the commands
FROM --platform=${OS}/${ARCH} golang:1.24 AS builder
ARG OS
ARG ARCH
WORKDIR /usr/src/app
COPY . .
RUN OS=${OS} ARCH=${ARCH} make

# Copy server to /usr/local/bin
FROM --platform=${OS}/${ARCH} debian:trixie-slim
ARG OS
ARG ARCH
ARG SOURCE
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/src/app/build/* /usr/local/bin/

# Labels
LABEL org.opencontainers.image.source=${SOURCE}

# Entrypoint
ENTRYPOINT [ "/usr/local/bin/api" ]
