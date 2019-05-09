FROM busybox:1.30

# Copy built binary.
COPY ./vaingogh /bin/vaingogh

ENTRYPOINT ["vaingogh"]
