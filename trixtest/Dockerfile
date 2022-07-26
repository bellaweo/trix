FROM matrixdotorg/synapse:v1.63.1

COPY homeserver.db /data/homeserver.db
COPY homeserver.yaml /data/homeserver.yaml
COPY media_store/ /data/media_store/
COPY trix.meh.log.config /data/trix.meh.log.config
COPY trix.meh.signing.key /data/trix.meh.signing.key

RUN cd / && chown -R 991:991 /data
