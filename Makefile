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
SCRIPTS_DIR := scripts
STORAGE_DIR := storage
CMD_DIR := cmd/raftnode
run: proto
	$(SCRIPTS_DIR)/simple_example.sh $(STORAGE_DIR) $(CMD_DIR)
