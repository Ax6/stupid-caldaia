FROM node AS build

WORKDIR /app

COPY package.json ./
RUN npm install
COPY . ./
RUN npm run build

ENV PUBLIC_SERVER_HOST=controller

EXPOSE 4173

CMD ["npm", "run", "host"]