.PHONY: proto_server
proto_server:
	@protoc -I./pkg/server \
		--go_out ./pkg/server \
		--go_opt=paths=source_relative \
		--go-grpc_out ./pkg/server \
		--go-grpc_opt=paths=source_relative \
		./pkg/server/*.proto

