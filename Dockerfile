FROM blufor/lightimage
RUN mkdir -p /etc/metcap
ENV REDIS_HOST redis
ENV REDIS_PORT 6379
ENV REDIS_DB 0
ENV REDIS_QUEUE metrics
ENV ES_URL http://es:9200/
ENV ES_INDEX metrics
ENV ES_TYPE raw
ENV WRITER_CONCURRENCY 2
ENV INFLUX_PORT 8001
ENV GRAPHITE_PORT 8002
COPY bin/metcap /bin/metcap
COPY bin/metcap-docker /bin/metcap-docker
VOLUME /etc/metcap
CMD [ "/bin/metcap-docker" ]
