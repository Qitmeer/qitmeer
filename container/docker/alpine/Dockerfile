FROM alpine:3.12
MAINTAINER jin

ENV BUILDDIR=/qitmeer

RUN apk add --no-cache curl bash jq curl && \
    rm -rf /var/cache/apk/*

COPY ./build/ $BUILDDIR

ENTRYPOINT ["/qitmeer/launch"]

