GO_BIN_FILES=es-convert-to-dads.go
GO_BIN_CMDS=es-convert-to-dads
#for race CGO_ENABLED=1
#GO_ENV=CGO_ENABLED=1
GO_ENV=CGO_ENABLED=0
#GO_BUILD=go build -ldflags '-s -w' -race
GO_BUILD=go build -ldflags '-s -w'
GO_INSTALL=go install -ldflags '-s'
GO_FMT=gofmt -s -w
GO_LINT=golint -set_exit_status
GO_VET=go vet
GO_CONST=goconst
GO_IMPORTS=goimports -w
GO_USEDEXPORTS=usedexports
GO_ERRCHECK=errcheck -asserts -ignore '[FS]?[Pp]rint*'
BINARIES=es-convert-to-dads
STRIP=strip

all: check ${BINARIES}

es-convert-to-dads: es-convert-to-dads.go
	 ${GO_ENV} ${GO_BUILD} -o es-convert-to-dads es-convert-to-dads.go

fmt: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_FMT}"

lint: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_LINT}"

vet: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_VET}"

imports: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_IMPORTS}"

usedexports: ${GO_BIN_FILES}
	${GO_USEDEXPORTS} ./...

errcheck: ${GO_BIN_FILES}
	${GO_ERRCHECK} ./...

check: fmt lint imports vet usedexports errcheck

install: check ${BINARIES}
	${GO_INSTALL} ${GO_BIN_CMDS}

strip: ${BINARIES}
	${STRIP} ${BINARIES}

clean:
	rm -f ${BINARIES}
