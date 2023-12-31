FROM node:alpine AS build

ENV NPM_CONFIG_CACHE=/usr/src/app/.npm-cache

WORKDIR /app
COPY package.json /app

RUN npm config set registry https://registry.npmjs.org/
RUN npm install --verbose

COPY . ./
ENV PUBLIC_SERVER_HOST=controller
ENV PUBLIC_CLIENT_HOST=192.168.1.112
RUN npm run build

EXPOSE 4173

CMD ["npm", "run", "host"]