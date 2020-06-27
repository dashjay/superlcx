.PHONY: plugins

GO = go
SOURCE = cmd/main.go
TARGET = superlcx

build:
	${GO} build -o ${TARGET} ${SOURCE}

slim:
	${GO} build -o temp ${SOURCE}
	upx -9 -q -o ${TARGET}_upx temp
	rm temp

plugins:
	./build_all_plugins.sh
