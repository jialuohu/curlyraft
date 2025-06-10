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
LOG_DIR := $(STORAGE_DIR)/log
run: proto
	$(SCRIPTS_DIR)/simple_example.sh $(STORAGE_DIR) $(CMD_DIR) $(LOG_DIR)

runA:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeA \
      --addr  localhost:21001 \
      --peers nodeB/localhost:21002,nodeC/localhost:21003 \
      --data  "$(STORAGE_DIR)/21001"
#      2>&1 | tee "$(LOG_DIR)/21001.log"

runB:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeB \
      --addr  localhost:21002 \
      --peers nodeA/localhost:21001,nodeC/localhost:21003 \
      --data  "$(STORAGE_DIR)/21002"
#      2>&1 | tee "$(LOG_DIR)/21002.log"

runC:
	go run "$(CMD_DIR)/main.go" \
      --id    nodeC \
      --addr  localhost:21003 \
      --peers nodeA/localhost:21001,nodeB/localhost:21002 \
      --data  "$(STORAGE_DIR)/21003"
#      2>&1 | tee "$(LOG_DIR)/21003.log"
