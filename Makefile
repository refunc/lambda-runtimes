SHELL := $(shell which bash) # ensure bash is used
export BASH_ENV=scripts/common

GOOS := $(shell eval $$(go env); echo $${GOOS})
ARCH := $(shell eval $$(go env); echo $${GOARCH})

runtimes := python3.7 python3.8

runner: bin/$(GOOS)/runner

bin/$(GOOS)/runner: main.go
	@echo GOOS=$(GOOS)
	CGO_ENABLED=0 go build \
	-tags netgo -installsuffix netgo \
	-ldflags "-s -w $(LD_FLAGS)" \
	-a \
	-o $@ \
	*.go

$(runtimes):
	@cd $@; \
	docker build \
	--build-arg https_proxy="$${HTTPS_RPOXY}" \
	--build-arg http_proxy="$${HTTP_RPOXY}" \
	-t refunc/lambda:$@ .
.PHONY: $(runtimes)

ifneq ($(GOOS),linux)
shell-%:
	@export GOOS=linux; make $@
else
shell-%: % bin/linux/runner
	docker run --rm -it -v $$(pwd)/examples/$*:/var/task -v $$(pwd)/bin/linux/runner:/bin/runner --entrypoint=/bin/bash refunc/lambda:$*
endif

ifneq ($(GOOS),linux)
test-%:
	@export GOOS=linux; make $@
else
test-%: % bin/linux/runner
	docker run --rm -it -v $$(pwd)/examples/$*:/var/task -v $$(pwd)/bin/linux/runner:/bin/runner --entrypoint=/bin/bash refunc/lambda:$* ./test.sh
endif

push: $(runtimes)
	@for img in $(runtimes); do \
	log_info "pushing refunc/lambda:$${img}" ; \
	docker push refunc/lambda:$${img}; \
	done