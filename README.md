# k8s-deployment-watcher
## Getting started

make build

## To Build protobuf
### Install protoc, protoc-gen-go 
brew install protoc-gen-go
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc

### build protobuf files
cd proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative podstat.proto