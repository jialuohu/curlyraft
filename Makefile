.PHONY: all
all:
	@echo usage:
	@echo "  make proto"

.PHONY: proto
PROTO_DIR   := proto
PROTO_GEN_DIR := internal/proto
PROTO_FILES := $(PROTO_DIR)/raftcomm/*.proto \
			   $(PROTO_DIR)/gateway/*.proto
proto:
	protoc \
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
LOG_DIR := $(STORAGE_DIR)/log
run: proto
	$(SCRIPTS_DIR)/simple_example.sh $(STORAGE_DIR) $(CMD_DIR) $(LOG_DIR)

ra:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeA \
      --addr  127.0.0.1:21001 \
      --peers nodeB/127.0.0.1:21002,nodeC/127.0.0.1:21003 \
      --data  "$(STORAGE_DIR)/21001"

rb:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeB \
      --addr  127.0.0.1:21002 \
      --peers nodeA/127.0.0.1:21001,nodeC/127.0.0.1:21003 \
      --data  "$(STORAGE_DIR)/21002"

rc:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeC \
      --addr  127.0.0.1:21003 \
      --peers nodeA/127.0.0.1:21001,nodeB/127.0.0.1:21002 \
      --data  "$(STORAGE_DIR)/21003"
