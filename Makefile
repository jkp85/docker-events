SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=stats

VERSION=0.0.1

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build -o ${BINARY} *.go

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
