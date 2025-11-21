module github.com/mpataki/go-job-queue/service

go 1.25

require (
	connectrpc.com/connect v1.19.1
	connectrpc.com/grpcreflect v1.3.0
	github.com/mpataki/go-job-queue/proto v0.0.0
	github.com/redis/go-redis/v9 v9.17.0
	golang.org/x/net v0.34.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)

replace github.com/mpataki/go-job-queue/proto => ../proto
