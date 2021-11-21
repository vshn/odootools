FROM docker.io/library/alpine:3.15

ENTRYPOINT ["odootools"]

RUN apk add --update --no-cache \
    bash \
    ca-certificates \
    curl

COPY odootools /usr/local/bin/

USER 65532
