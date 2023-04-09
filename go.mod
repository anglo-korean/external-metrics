module github.com/anglo-korean/{{ .Name }}

go 1.19

require (
	github.com/anglo-korean/logger-go v0.0.0-20221108210025-93b7a8e9ac30
	github.com/anglo-korean/protobuf/types/go v0.0.0-20221203104644-445d84fc1b98
	github.com/confluentinc/confluent-kafka-go v1.9.2
	google.golang.org/protobuf v1.28.1
)

require (
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
)
