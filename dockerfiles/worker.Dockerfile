FROM golang:1.21-alpine3.19 AS build

WORKDIR /app
COPY go.mod go.work go.work.sum /app/
COPY controller/go.mod controller/go.sum /app/controller/
COPY lettore/go.mod lettore/go.sum /app/lettore/

# This ensures that the dependencies are fetched
RUN go mod download

COPY controller /app/controller/
COPY lettore /app/lettore/

WORKDIR /app/lettore
RUN go build -o /lettore

FROM alpine:3.19 AS run

COPY --from=build /lettore /lettore
COPY config.json /config.json
ENV CONFIG_PATH=/config.json

EXPOSE 8080

ENTRYPOINT ["/lettore"]