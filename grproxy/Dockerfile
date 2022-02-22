FROM scratch
FROM golang:alpine
RUN apk update && \
        apk --no-cache add make && \
        apk --no-cache add git 
ADD /src/grproxy /src/grproxy
ENV GOPATH=/src
RUN go get github.com/samuel/go-zookeeper/zk
WORKDIR /
COPY Makefile Makefile
RUN make
WORKDIR /
CMD ["./bin/grproxy"]
