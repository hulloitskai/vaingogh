FROM alpine:3.9

# Install system dependencies.
RUN apk add --update ca-certificates

# Copy built binary.
ENV PROGRAM=vaingogh
COPY ./dist/${PROGRAM} /bin/${PROGRAM}

# Configure env and exposed ports.
ENV GOENV=production
EXPOSE 3000

ENTRYPOINT $PROGRAM serve
