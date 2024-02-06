FROM node:alpine AS build

WORKDIR /app
COPY package.json /app

RUN npm config set registry https://registry.npmjs.org/
RUN npm install --verbose

COPY src /app/src
COPY *config* /app/
ENV PUBLIC_SERVER_HOST=controller
ENV PUBLIC_CLIENT_HOST=192.168.1.123
RUN npm run build

EXPOSE 4173

CMD ["npm", "run", "host"]