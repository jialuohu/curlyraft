.PHONY: all
all:
	@echo usage:
	@echo "  make proto"

.PHONY: proto
PROTO_DIR   := proto
PROTO_GEN_DIR := internal/proto
PROTO_FILES := $(PROTO_DIR)/*.proto
proto:
	@protoc \
		-I $(PROTO_DIR) \
		--go_out=$(PROTO_GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

.PHONY: run
run: proto
	@go run cmd/raftnode/main.go
