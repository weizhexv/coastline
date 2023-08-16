#-alpine：指代镜像比较小
# FROM public.ecr.aws/bitnami/golang:1.19
FROM public.ecr.aws/docker/library/golang:1.19 as builder

LABEL stage=gobuilder

ARG GIT_USER
ARG GIT_PWD

# 环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    GOPRIVATE=dghire.com \
    CGO_ENABLED=0 \
    GOOS=linux \
    GODEBUG="http2debug=1" \
    GOARCH=amd64

WORKDIR /application

RUN git config --global url."https://$GIT_USER:$GIT_PWD@git-codecommit.eu-central-1.amazonaws.com/".insteadOf "https://git-codecommit.eu-central-1.amazonaws.com/"

#COPY go.mod , go.sum and download the dependencied
COPY . .
RUN go mod download

RUN go build -ldflags "-s -w" -o /application/build/qa-coastline coastline.go

FROM public.ecr.aws/ubuntu/ubuntu:22.04_stable
# FROM public.ecr.aws/docker/library/golang:1.19-alpine

WORKDIR /target

# 复制编译后的程序
COPY --from=builder /application/build/qa-coastline /target/qa-coastline
COPY --from=builder /application/vconfig/ /target/vconfig
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8080
ENTRYPOINT ["/target/qa-coastline"]
