FROM node:20-alpine AS build
WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
ARG VITE_API_BASE=/api
ARG VITE_WS_BASE=/api/v1/ws
ARG VITE_TILE_URL=/tiles/{z}/{x}/{y}.png
ENV VITE_API_BASE=$VITE_API_BASE \
    VITE_WS_BASE=$VITE_WS_BASE \
    VITE_TILE_URL=$VITE_TILE_URL
COPY package.json ./
RUN npm install --registry=https://registry.npmmirror.com
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache wget
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
