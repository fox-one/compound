# 源镜像
FROM alpine:latest

RUN apk --no-cache add tzdata

# 设置工作目录
WORKDIR /app

# Add config files
ADD config/config.yaml config/config.yaml

# 将二进制可执行文件加入到docker容器中
ADD compound compound

# # 暴露端口
EXPOSE 80

# # 最终运行docker的命令，以 WORKDIR 为基准
ENTRYPOINT  ["./compound"]
