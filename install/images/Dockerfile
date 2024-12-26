FROM alpine:3.21

# install dependencies.
RUN apk add jq vim busybox-extras libc6-compat redis

# install mongodb with old package cause it doesn’t exist for Alpine after version 3.9.
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/main' >> /etc/apk/repositories
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/community' >> /etc/apk/repositories
RUN apk add mongodb && rm -f /usr/bin/mongod /usr/bin/mongos

WORKDIR /bk-nodeman/bin
COPY bk-nodeman-backend .
