FROM matrixdotorg/synapse:latest

COPY homeserver.yaml /data/homeserver.yaml
COPY localhost.signing.key /data/localhost.signing.key
COPY localhost.log.config /data/localhost.log.config
RUN cd / && chown -R 991:991 /data
