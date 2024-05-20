generate_grpc_code:
	mkdir -p emailService
	protoc \
		--proto_path=. \
		--proto_path=$(PROTOC_INC_PATH) \
		--go_out=emailService \
		--go_opt=paths=source_relative \
		--go-grpc_out=emailService \
		--go-grpc_opt=paths=source_relative \
		emailService.proto
