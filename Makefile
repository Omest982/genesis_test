generate_grpc_code:
	protoc \
	--go_out=emailService \
	--go_opt=paths=source_relative \
	--go-grpc_out=emailService \
	--go-grpc_opt=paths=source_relative \
	emailService.proto