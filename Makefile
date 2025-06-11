.PHONY: all
all:
	@echo usage:
	@echo "  make proto"
	@echo "  make clean"

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

.PHONY: clean
clean:
	rm -rf ./storage

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
      --peers nodeB/127.0.0.1:21002,nodeC/127.0.0.1:21003,nodeD/127.0.0.1:21004,nodeE/127.0.0.1:21005 \
      --data  "$(STORAGE_DIR)/21001"

rb:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeB \
      --addr  127.0.0.1:21002 \
      --peers nodeA/127.0.0.1:21001,nodeC/127.0.0.1:21003,nodeD/127.0.0.1:21004,nodeE/127.0.0.1:21005 \
      --data  "$(STORAGE_DIR)/21002"

rc:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeC \
      --addr  127.0.0.1:21003 \
      --peers nodeA/127.0.0.1:21001,nodeB/127.0.0.1:21002,nodeD/127.0.0.1:21004,nodeE/127.0.0.1:21005 \
      --data  "$(STORAGE_DIR)/21003"

rd:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeD \
      --addr  127.0.0.1:21004 \
      --peers nodeA/127.0.0.1:21001,nodeB/127.0.0.1:21002,nodeC/127.0.0.1:21003,nodeE/127.0.0.1:21005 \
      --data  "$(STORAGE_DIR)/21004"

re:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeE \
      --addr  127.0.0.1:21005 \
      --peers nodeA/127.0.0.1:21001,nodeB/127.0.0.1:21002,nodeC/127.0.0.1:21003,nodeD/127.0.0.1:21004 \
      --data  "$(STORAGE_DIR)/21005"
