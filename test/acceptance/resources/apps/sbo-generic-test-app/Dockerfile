FROM alpine:3.12

RUN apk update && \
    apk add --no-cache \
    lighttpd \
    bash \
    jq && \
    rm -rf /var/cache/apk/*

RUN chmod -R go+rwx /run /var

COPY conf/* /etc/lighthttpd/
COPY start_http /bin/
COPY env.sh /bin/

EXPOSE 8080

ENTRYPOINT ["/bin/start_http"]
