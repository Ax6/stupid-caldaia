FROM node:alpine3.18 AS build

WORKDIR /app

COPY package.json ./
COPY package-lock.json ./

RUN apk --no-cache --virtual build-dependencies add \
        python \
        make \
        g++

RUN npm i
COPY . ./
RUN npm run build

FROM nginx:alpine3.18 AS runtime

COPY --from=build /app /usr/share/nginx/html
