FROM node:20
WORKDIR /app/binance-frontend
RUN npm install -g yarn --force
COPY binance-frontend/package*.json ./
COPY binance-frontend/yarn.lock ./
RUN yarn install

COPY binance-frontend/ .
EXPOSE 8080
CMD ["yarn", "dev", "--host", "0.0.0.0", "--port", "8080"]