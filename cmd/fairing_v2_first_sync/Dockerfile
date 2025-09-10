FROM gcr.io/distroless/static-debian11

# 将工作目录设置为根目录
WORKDIR /

# 将我们在本地编译好的可执行文件（我们将其命名为 'server'）复制到镜像的根目录
COPY server /server

# 设置容器启动时默认执行的命令
CMD ["/server"]