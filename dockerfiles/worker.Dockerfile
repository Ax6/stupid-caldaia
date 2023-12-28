FROM alpine:3.19 AS run

RUN apk update && apk add i2c-tools

COPY --from=executables lettore /lettore
COPY config_prod.json /config.json
ENV CONFIG_PATH=/config.json

ENTRYPOINT ["/lettore"]