FROM docker.io/library/alpine:3.16

ENTRYPOINT ["odootools"]

RUN apk add --update --no-cache \
    bash \
    ca-certificates \
    curl \
    tzdata

COPY odootools /usr/local/bin/

USER 65532
