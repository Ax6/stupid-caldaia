FROM golang AS build

WORKDIR /app
COPY . .
RUN go build -o /controller

FROM alpine AS run

COPY --from=build /controller /controller

EXPOSE 8080

ENTRYPOINT ["/controller"]