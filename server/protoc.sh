#!/bash/bin

set -e

protoc -I ./proto --go_out=plugins=grpc:./ vote.proto