FROM alpine:3.19 AS run

COPY --from=executables lettore /lettore
COPY config-prod.json /config.json
ENV CONFIG_PATH=/config.json

EXPOSE 8080

ENTRYPOINT ["/lettore"]