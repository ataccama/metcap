FROM blufor/lightimage
RUN mkdir -p /etc/metrics-capacitor
ENV METCAP_REDIS "127.0.0.1:6379"
ENV METCAP_ELASTIC "127.0.0.1:9200"
COPY bin/metrics-capacitor /bin/metrics-capacitor
COPY metrics-capacitor-docker /bin/metrics-capacitor-docker
VOLUME [ "/etc/metrics-capacitor" ]
CMD [ "/bin/metrics-capacitor-docker" ]
