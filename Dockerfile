FROM alpine:latest
RUN mkdir -p /etc/metcap /go && \
    apk add --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community/ git go gcc musl-dev && \
    export GOPATH=/go && \
    go get -v github.com/blufor/metcap github.com/blufor/metcap/cmd/metcap && \
    go build -o /bin/metcap github.com/blufor/metcap/cmd/metcap && \
    apk del git go gcc musl-dev && \
    rm -rf /tmp/* /go /usr/lib/go /var/log
ENV REPORT_EVERY 10s
ENV TRANSPORT_TYPE channel
ENV TRANSPORT_SIZE 1000000
ENV REDIS_URL tcp://127.0.0.1:6379|0
ENV REDIS_QUEUE metrics
ENV REDIS_TIMEOUT 5
ENV REDIS_WAIT 1
ENV REDIS_RETRIES 3
ENV REDIS_CONNECTIONS 100
ENV AMQP_URL amqp://guest:guest@127.0.0.1:5672/
ENV AMQP_TIMEOUT 5
ENV AMQP_TAG default
ENV AMQP_WORKERS 1
ENV LISTENER_DECODERS 2
ENV LISTENER_GRAPHITE_RULES '^STRESS\.host|||-.-.host.-.-.1.2+'
ENV WRITER_ES_URL http://127.0.0.1:9200
ENV WRITER_TIMEOUT 10
ENV WRITER_INDEX metrics
ENV WRITER_DOC_TYPE raw
ENV WRITER_CONCURRENCY 2
ENV WRITER_BULK_MAX 5000
ENV WRITER_BULK_WAIT 10s
COPY ./bin/docker-metcap /bin/docker-metcap
VOLUME /etc/metcap /tmp
ENTRYPOINT [ "/bin/docker-metcap" ]
