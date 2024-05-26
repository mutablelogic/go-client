ARG OS
ARG ARCH

# Run makefile to build all the commands
FROM --platform=${OS}/${ARCH} golang:1.21-bullseye AS builder
ARG OS
ARG ARCH
WORKDIR /usr/src/app
COPY . .
RUN OS=${OS} ARCH=${ARCH} make

# Copy server to /usr/local/bin
FROM --platform=${OS}/${ARCH} debian:bullseye-slim
ARG OS
ARG ARCH
ARG SOURCE
RUN apt update && apt install -y ca-certificates
COPY --from=builder /usr/src/app/build/* /usr/local/bin/
COPY etc/entrypoint.sh .

# Labels
LABEL org.opencontainers.image.source=${SOURCE}

# Entrypoint when running the server
ENTRYPOINT [ "/entrypoint.sh" ]
CMD [ "/usr/local/bin/api" ]
