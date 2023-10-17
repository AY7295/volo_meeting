FROM --platform=${TARGETPLATFORM} golang:alpine as builder
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

RUN go build -o /app/runner main.go

FROM --platform=${TARGETPLATFORM} alpine:latest

RUN echo "http://mirrors.aliyun.com/alpine/v3.18/main/" > /etc/apk/repositories
RUN apk update && \
    apk upgrade --no-cache && \
    apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo 'Asia/Shanghai' >/etc/timezone && \
    rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /app/runner /app/runner

ENTRYPOINT [ "/app/runner" ]
