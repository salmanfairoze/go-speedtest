include build/make/Makefile.auto

.DEFAULT_GOAL: all

.PHONY: all
all: build

.PHONY: build
build: $(BINARIES)

$(BIN_DIR)/%: $(SOURCES) $(MOD_DEP)
	go build $(LD_FLAGS) -o $@ $(@:$(BIN_DIR)/%=$(CMD_DIR)/%)/*.go

.PHONY: clean
clean:
	rm -fv $(BINARIES)
