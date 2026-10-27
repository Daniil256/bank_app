module customer-service

go 1.26.4

require (
	bank_app v0.0.0
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.22.0
	google.golang.org/grpc v1.84.0
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.33.0 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261026165753-2ffab32bce7e // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace bank_app => ../
