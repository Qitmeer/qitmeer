FROM alpine:3.9
MAINTAINER jin

ENV BUILDDIR=/nox

RUN apk add --no-cache curl bash jq curl && \
    rm -rf /var/cache/apk/*

COPY ./ $BUILDDIR

ENTRYPOINT ["/nox/launch"]

