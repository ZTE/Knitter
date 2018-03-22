OUT_DIR = _output/bin

export GOFLAGS

# all make sub options
default: help

help:
	@echo "Usage: make <target>"
	@echo
	@echo " 'build'          - Build all knitter related binaries(e.g. knitter-manager,knitter-agent,knitter-plugin,knitter-mornitor)"
	@echo " 'test-ut'        - Test knitter with unit test"
	@echo " 'test-e2e'       - Test knitter with e2e test"
	@echo " 'clean'          - Clean all output artifacts"
	@echo " 'verify'         - Execute the source code verification tools(e.g. gofmt,lint,govet)"
	@echo " 'install-extra'  - Install tools used by verify(e.g. gometalinter)"

# Ideally not needed this section
check-gopath:
ifndef GOPATH
        $(error GOPATH is not set)
endif
.PHONY: check-gopath

# Build code.
#
# Args:
#   GOFLAGS: Extra flags to pass to 'go' when building.
#
# Example:
#         make build
#         make all
all build: knitter-manager knitter-agent knitter-plugin knitter-monitor
.PHONY: all build

# Build knitter-plugin
#
# Example:
#        make knitter-plugin
knitter-plugin:
	hack/build/plugin.sh build
	mkdir -p ${OUT_DIR}
	mv ./knitter-plugin/knitter-plugin   ${OUT_DIR}
.PHONY: knitter-plugin

# Build knitter-manager
#
# Example:
#         make knitter-manager
knitter-manager:
	hack/build/manager.sh build
	mkdir -p ${OUT_DIR}
	mv ./knitter-manager/knitter-manager   ${OUT_DIR}
.PHONY: knitter-manager

# Build knitter-monitor
#
# Example:
#         make knitter-monitor
knitter-monitor:
	hack/build/monitor.sh build
	mkdir -p ${OUT_DIR}
	mv ./knitter-monitor/knitter-monitor   ${OUT_DIR}
.PHONY: knitter-monitor

# Build knitter-agent
#
# Example:
#         make knitter-agent
knitter-agent:
	hack/build/agent.sh build
	mkdir -p ${OUT_DIR}
	mv ./knitter-agent/knitter-agent   ${OUT_DIR}
.PHONY: knitter-agent

# Lint knitter code files. note that this lint process handled by gometalinter tools.
# link here (github.com/alecthomas/gometalinter)
# If users only need simple lint process, please run command 'make golint'.
# Example:
#         make lint
lint: check-gopath
	@echo "checking lint"
	hack/verify-lint.sh

# Lint knitter code files with golint tool.
#
# Example:
#         make golint
golint: check-gopath
	@echo "checking golint"
	hack/verify-golint.sh

# Format knitter code files with gofmt tool.
# 
# Example:
#         make gofmt
gofmt:
	@echo "checking gofmt"
	hack/verify-gofmt.sh

# Static check knitter code files 
#
# Example:
#         make govet
govet:
	@echo "checking govet"
	hack/verify-govet.sh

# Verify whether code is properly organized.
#
# Example:
#         make verify
verify: gofmt govet
.PHONY: verify

# strict-verify do more strict verify process
#
# Example:
#         make strict-verify
strict-verify:gofmt lint govet
.PHONY:strict-verify

# Run verify and test process
#
# Example
# make verify
check: verify test-ut	
.PHONY: check

# Run unit tests
# 
# Example:
# make test-ut
test-ut:
	go test -timeout=20m -race ./pkg/... ./knitter-agent/... ./knitter-manager/... ./knitter-monitor/... ./knitter-plugin/... $(BUILD_TAGS) $(GO_LDFLAGS) $(GO_GCFLAGS)
.PHONY: test-ut

# Run coverage checking
#
# Example:
#make test-cover
test-cover:
	PATH="$HOME/gopath/bin:$PATH"
	./hack/cover.sh --coveralls
.PHTONY: test-cover

install-extra:install-gometalinter
.PHONY: install-extra

# install gometailinter tool
#
# Example:
# make install-gometalinter
install-gometalinter:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
.PHONY: install-gometalinter

# Deploy a dind kubernetes cluser
# note: Leverage Mirantis kubeadm-dind-cluster
# Example:
# make deploy-k8s
deploy-k8s:
	./hack/dind-cluster-v1.8.sh up
.PHONY: deploy-k8s

# Run knitter e2e test case
# Example:
# make test-e2e
# NOTE THAT: knitter test-e2e is experimental now.
test-e2e:
	./hack/run-robotframe-e2e.sh
.PHONY: test-e2e

# Just for test
probe-test:
	./hack/probe-test.sh
.PHONY: probe-test

# Remove all build artifacts.
#
# Example:
#   make clean
clean:
	rm -rf ./_output
	rm -rf .cover
.PHONY: clean


