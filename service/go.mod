module github.com/mpataki/go-job-queue/service

go 1.25

require (
	connectrpc.com/connect v1.19.1
	connectrpc.com/grpcreflect v1.3.0
	github.com/mpataki/go-job-queue/proto v0.0.0
	golang.org/x/net v0.34.0
)

require (
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)

replace github.com/mpataki/go-job-queue/proto => ../proto
