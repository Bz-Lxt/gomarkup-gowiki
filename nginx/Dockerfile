FROM node:22-alpine AS fe
WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
COPY frontend-user/package.json frontend-user/package-lock.json* ./
RUN npm install --registry=https://registry.npmmirror.com
COPY frontend-user/ ./
RUN npm run build

FROM nginx:1.27-alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache wget tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo Asia/Shanghai > /etc/timezone
ENV TZ=Asia/Shanghai
COPY --from=fe /app/dist /usr/share/nginx/html
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
