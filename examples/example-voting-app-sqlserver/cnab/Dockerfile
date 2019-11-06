FROM alpine:latest

RUN apk update && apk add make

# This is a modified version of the main Dockerfile for compose.
# The upstream version sets an entrypoint.
# See https://github.com/docker/compose/blob/master/Dockerfile.run

ENV GLIBC 2.27-r0
ENV COMPOSE 1.22.0
ENV DOCKERBINS_SHA 1270dce1bd7e1838d62ae21d2505d87f16efc1d9074645571daaefdfd0c14054

RUN apk update && apk add --no-cache openssl ca-certificates curl libgcc && \
    curl -fsSL -o /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub && \
    curl -fsSL -o glibc-$GLIBC.apk https://github.com/sgerrand/alpine-pkg-glibc/releases/download/$GLIBC/glibc-$GLIBC.apk && \
    apk add --no-cache glibc-$GLIBC.apk && \
    ln -s /lib/libz.so.1 /usr/glibc-compat/lib/ && \
    ln -s /lib/libc.musl-x86_64.so.1 /usr/glibc-compat/lib && \
    ln -s /usr/lib/libgcc_s.so.1 /usr/glibc-compat/lib && \
    curl -fsSL -o dockerbins.tgz "https://download.docker.com/linux/static/stable/x86_64/docker-17.12.1-ce.tgz" && \
    echo "${DOCKERBINS_SHA}  dockerbins.tgz" | sha256sum -c - && \
    tar xvf dockerbins.tgz docker/docker --strip-components 1 && \
    mv docker /usr/local/bin/docker && \
    chmod +x /usr/local/bin/docker && \
    curl -fsSL -o docker-compose-bin "https://github.com/docker/compose/releases/download/$COMPOSE/docker-compose-Linux-x86_64" && \
    mv docker-compose-bin /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose && \
    rm dockerbins.tgz /etc/apk/keys/sgerrand.rsa.pub glibc-$GLIBC.apk && \
    apk del curl

COPY app /cnab/app
COPY Dockerfile /cnab/Dockerfile

CMD ["/cnab/app/run"]