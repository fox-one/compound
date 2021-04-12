# 源镜像
FROM golang:1.16 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN make build

FROM frolvlad/alpine-glibc:latest

# 设置工作目录
WORKDIR /app

COPY --from=builder /app/builds/compound .

# # 暴露端口
EXPOSE 80

# # 最终运行docker的命令，以 WORKDIR 为基准
ENTRYPOINT  ["./compound"]
