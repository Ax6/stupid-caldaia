FROM node:alpine AS build
ARG LATITUDE
ARG LONGITUDE
ARG TIMEZONE

WORKDIR /app
COPY package.json /app

RUN npm config set registry https://registry.npmjs.org/
RUN npm install --verbose

COPY src /app/src
COPY *config* /app/
ENV PUBLIC_SERVER_HOST=controller
ENV PUBLIC_CLIENT_HOST=192.168.1.123
ENV LATITUDE=${LATITUDE}
ENV LONGITUDE=${LONGITUDE}
ENV TIMEZONE=${TIMEZONE}
RUN npm run build

EXPOSE 4173

CMD ["npm", "run", "host"]