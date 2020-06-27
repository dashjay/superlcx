.PHONY: plugins

GO = go
SOURCE = cmd/main.go
TARGET = superlcx
GOOS = darwin
ARCH = amd64

build:
	GOOS=${GOOS} ARCH=${ARCH} ${GO} build -o ${TARGET} ${SOURCE}

slim:
	${GO} build -o temp ${SOURCE}
	upx -9 -q -o ${TARGET}_upx temp
	rm temp

plugins:
	./build_all_plugins.sh
