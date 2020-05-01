FROM golang:1.13.10-buster

ADD ./protoc /bin/protoc
ENV GOPATH /.go
RUN apt-get -y install git

RUN go get -v google.golang.org/grpc
RUN go get -v github.com/golang/protobuf/protoc-gen-go
RUN go get -v github.com/go-sql-driver/mysql
RUN go get -v github.com/BurntSushi/toml

CMD ["go", "run", "main.go"]
