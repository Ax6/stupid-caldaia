FROM node:21-alpine AS build

ENV NPM_CONFIG_CACHE=/usr/src/app/.npm-cache

WORKDIR /tmp
ADD package.json /tmp/
RUN npm config set registry https://registry.npmjs.org/
RUN npm install --verbose
RUN cp -a /tmp/node_modules /app/

WORKDIR /app
COPY . ./
ENV PUBLIC_SERVER_HOST=controller
RUN npm run build

EXPOSE 4173

CMD ["npm", "run", "host"]