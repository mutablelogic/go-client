ARG OS
ARG ARCH

# Run makefile to build all the commands
FROM --platform=${OS}/${ARCH} golang:1.22 AS builder
ARG OS
ARG ARCH
WORKDIR /usr/src/app
COPY . .
RUN OS=${OS} ARCH=${ARCH} make

# Copy server to /usr/local/bin
FROM --platform=${OS}/${ARCH} debian:bookworm-slim
ARG OS
ARG ARCH
ARG SOURCE
RUN apt update && apt install -y ca-certificates
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
COPY --chmod=755 etc/entrypoint.sh .

# Install the api commands
RUN api install

# Labels
LABEL org.opencontainers.image.source=${SOURCE}

# Entrypoint when running the server
ENTRYPOINT [ "/entrypoint.sh" ]
CMD [ "/usr/local/bin/api" ]
