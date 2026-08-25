FROM golang:1.24-alpine AS build
WORKDIR /src
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache git ca-certificates
ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOTOOLCHAIN=auto
COPY go.mod go.sum ./
COPY . .
RUN go build -mod=mod -o /out/godruid ./cmd/godruid

FROM alpine:3.20
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates wget tzdata \
    && adduser -D -u 65532 app
WORKDIR /app
COPY --from=build /out/godruid /app/godruid
USER 65532
ENV TZ=Asia/Shanghai
EXPOSE 8080
ENTRYPOINT ["/app/godruid"]
