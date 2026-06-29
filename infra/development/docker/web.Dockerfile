FROM node:20-alpine

WORKDIR /app

COPY web/package*.json ./

RUN npm install

COPY web ./

RUN npm run build

EXPOSE 3000

CMD ["npm", "start"]

#COPY web/package*.json ./ before COPY web ./ is the standard Docker layer-caching trick. npm install is the slowest step in this file. By copying only package.json/package-lock.json first and running npm install before copying the rest of the source, Docker can reuse the cached npm install layer for any change that doesn't touch dependencies — only RUN npm run build and the layers after it re-run.

