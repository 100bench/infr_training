module github.com/100bench/infr_training/notify_api

go 1.24.3

replace github.com/100bench/infr_training/pkg => ../pkg

require (
	github.com/100bench/infr_training/pkg v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.78.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
