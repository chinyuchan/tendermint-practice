builtin:
	go build -o ./build/kvstore-builtin ./kvstore/builtin/main.go

external:
	go build -o ./build/kvstore-external ./kvstore/external/main.go

all: builtin external

.PHONY: buildin separate

