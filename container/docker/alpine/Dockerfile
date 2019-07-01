FROM alpine:3.9
MAINTAINER jin

ENV BUILDDIR=/qitmeer

RUN apk add --no-cache curl bash jq curl && \
    rm -rf /var/cache/apk/*

COPY ./build/ $BUILDDIR

ENTRYPOINT ["/qitmeer/launch"]

