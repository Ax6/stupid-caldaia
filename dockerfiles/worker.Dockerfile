FROM alpine:3.19 AS run

COPY --from=executables lettore /lettore
COPY config-prod.json /config.json
ENV CONFIG_PATH=/config.json

RUN apk update && apk add i2c-tools

RUN usermod -aG plugdev $USER

ENTRYPOINT ["/lettore"]