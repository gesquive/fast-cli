#
#  Makefile
#
#  The kickoff point for all project management commands.
#

GOCC := go

# Program version
VERSION := $(shell git describe --always --tags)

# Binary name for bintray
BIN_NAME=fast-cli

# Project owner for bintray
OWNER=gesquive

# Project name for bintray
PROJECT_NAME=fast-cli

# Project url used for builds
# examples: github.com, bitbucket.org
REPO_HOST_URL=github.com

# Grab the current commit
GIT_COMMIT=$(shell git rev-parse HEAD)

# Check if there are uncommited changes
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)

# Use a local vendor directory for any dependencies; comment this out to
# use the global GOPATH instead
# GOPATH=$(PWD)

INSTALL_PATH=$(GOPATH)/${REPO_HOST_URL}/${OWNER}/${PROJECT_NAME}

FIND_DIST:=find * -type d -exec

default: build

help:
	@echo 'Management commands for digitalocean-ddns:'
	@echo
	@echo 'Usage:'
	@echo '    make build    Compile the project'
	@echo '    make link     Symlink this project into the GOPATH'
	@echo '    make test     Run tests on a compiled project'
	@echo '    make install  Install binary'
	@echo '    make depends  Download dependencies'
	@echo '    make fmt      Reformat the source tree with gofmt'
	@echo '    make clean    Clean the directory tree'
	@echo '    make dist     Cross compile the full distribution'
	@echo

build:
	@echo "building ${OWNER} ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	${GOCC} build -ldflags "-X main.version=${VERSION} -X main.dirty=${GIT_DIRTY}" -o ${BIN_NAME}

install: build
	install -d ${DESTDIR}/usr/local/bin/
	install -m 755 ./${BIN_NAME} ${DESTDIR}/usr/local/bin/${BIN_NAME}

depends:
	${GOCC} get -u github.com/Masterminds/glide
	glide install

test:
	${GOCC} test ./...

clean:
	${GOCC} clean
	rm -f ./${BIN_NAME}.test
	rm -f ./${BIN_NAME}
	rm -rf ./dist

bootstrap-dist:
	${GOCC} get -u github.com/mitchellh/gox

build-all: bootstrap-dist
	gox -verbose \
	-ldflags "-X main.version=${VERSION} -X main.dirty=${GIT_DIRTY}" \
	-os="linux darwin windows " \
	-arch="amd64 386" \
	-output="dist/{{.OS}}-{{.Arch}}/{{.Dir}}" .

dist: build-all
	cd dist && \
	$(FIND_DIST) cp ../LICENSE {} \; && \
	$(FIND_DIST) cp ../README.md {} \; && \
	$(FIND_DIST) tar -zcf ${PROJECT_NAME}-${VERSION}-{}.tar.gz {} \; && \
	$(FIND_DIST) zip -r ${PROJECT_NAME}-${VERSION}-{}.zip {} \; && \
	cd ..

fmt:
	find . -name '*.go' -not -path './.vendor/*' -exec gofmt -w=true {} ';'

link:
	# relink into the go path
	if [ ! $(INSTALL_PATH) -ef . ]; then \
		mkdir -p `dirname $(INSTALL_PATH)`; \
		ln -s $(PWD) $(INSTALL_PATH); \
	fi


.PHONY: build help test install depends clean bootstrap-dist build-all dist fmt link
