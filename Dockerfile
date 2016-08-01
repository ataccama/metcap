FROM blufor/lightimage
RUN apk install --no-cache go
RUN mkdir -p /etc/metrics-capacitor
COPY app/bin/metrics-capacitor /bin/metrics-capacitor
COPY docker-metrics-capacitor /bin/docker-metrics-capacitor
VOLUME [ "/etc/metrics-capacitor" ]
CMD [ "/bin/docker-metrics-capacitor" ]
