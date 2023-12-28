FROM golang:1.21-alpine3.19 AS build

WORKDIR /app
COPY go.mod go.work go.work.sum /app/
COPY controller lettore /app/

WORKDIR /app/lettore
RUN go build -o /lettore

FROM alpine:3.19 AS run

COPY --from=build /lettore /lettore
COPY config.json /config.json
ENV CONFIG_PATH=/config.json

EXPOSE 8080

ENTRYPOINT ["/lettore"]