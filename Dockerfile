FROM golang:latest
MAINTAINER Bluek404 "i@bluek404.net"

# Build app
RUN go get github.com/Bluek404/aabbabab
RUN go install github.com/Bluek404/aabbabab

WORKDIR $GOPATH/src/github.com/Bluek404/aabbabab

EXPOSE 80
CMD ["aabbabab", "-host=:80", "-docker"]