FROM golang:1.25-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache git ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/gowiki ./cmd/server

FROM alpine:3.21
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata wget \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo Asia/Shanghai > /etc/timezone
ENV TZ=Asia/Shanghai \
    GOWIKI_UPLOAD_DIR=/data/uploads \
    GOWIKI_BLEVE_PATH=/data/bleve
WORKDIR /app
COPY --from=builder /out/gowiki /usr/local/bin/gowiki
RUN mkdir -p /data/uploads /data/bleve && chmod 777 /data/uploads /data/bleve
EXPOSE 8080
CMD ["gowiki"]
