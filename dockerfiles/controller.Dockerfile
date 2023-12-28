FROM alpine:3.19 AS run

COPY --from=executables controller /controller
COPY config.json /config.json
ENV CONFIG_PATH=/config.json

EXPOSE 8080

ENTRYPOINT ["/controller"]