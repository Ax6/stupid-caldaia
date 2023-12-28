FROM golang:1.21-alpine3.19 AS build

WORKDIR /app
COPY controller .
RUN go build -o /controller

FROM alpine:3.19 AS run

COPY --from=build /controller /controller
COPY config.json /config.json
ENV CONFIG_PATH=/config.json

EXPOSE 8080

ENTRYPOINT ["/controller"]