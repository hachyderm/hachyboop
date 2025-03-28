# =========================================================================== #
#            MIT License Copyright (c) 2022 Kris Nóva <kris@nivenly.com>      #
#                                                                             #
#                 ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓                 #
#                 ┃   ███╗   ██╗ ██████╗ ██╗   ██╗ █████╗   ┃                 #
#                 ┃   ████╗  ██║██╔═████╗██║   ██║██╔══██╗  ┃                 #
#                 ┃   ██╔██╗ ██║██║██╔██║██║   ██║███████║  ┃                 #
#                 ┃   ██║╚██╗██║████╔╝██║╚██╗ ██╔╝██╔══██║  ┃                 #
#                 ┃   ██║ ╚████║╚██████╔╝ ╚████╔╝ ██║  ██║  ┃                 #
#                 ┃   ╚═╝  ╚═══╝ ╚═════╝   ╚═══╝  ╚═╝  ╚═╝  ┃                 #
#                 ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛                 #
#                                                                             #
#                        This machine kills fascists.                         #
#                                                                             #
# =========================================================================== #

DATADIR := data

.PHONY: all
all: compile

version     ?=  0.0.1
target      ?=  hachyboop
org         ?=  hachyderm
authorname  ?=  Hachyderm Infrastructure Team
authoremail ?=  hachyderm@hachyderm.io
license     ?=  Apache 2.0
year        ?=  2025
copyright   ?=  Copyright (c) $(year)

.PHONY: compile
compile: ## Compile for the local architecture ⚙
	@echo "Compiling..."
	go build -ldflags "\
	-X 'github.com/$(org)/$(target).Version=$(version)' \
	-X 'github.com/$(org)/$(target).AuthorName=$(authorname)' \
	-X 'github.com/$(org)/$(target).AuthorEmail=$(authoremail)' \
	-X 'github.com/$(org)/$(target).Copyright=$(copyright)' \
	-X 'github.com/$(org)/$(target).License=$(license)' \
	-X 'github.com/$(org)/$(target).Name=$(target)'" \
	-o $(target) cmd/*.go

.PHONY: run
run: compile
	./hachyboop

.PHONY: container
container:
	docker build . -t $(target):latest

.PHONY: runcontainer
runcontainer: container
	docker run \
		-e HACHYBOOP_S3_ENDPOINT \
		-e HACHYBOOP_S3_BUCKET \
		-e HACHYBOOP_S3_PATH \
		-e HACHYBOOP_S3_ACCESS_KEY_ID \
		-e HACHYBOOP_S3_SECRET_ACCESS_KEY \
		-e HACHYBOOP_OBSERVER_ID \
		-e HACHYBOOP_OBSERVER_REGION \
		-e HACHYBOOP_RESOLVERS \
		-e HACHYBOOP_QUESTIONS \
		-e HACHYBOOP_LOCAL_RESULTS_PATH \
		-e BUNNYNET_MC_REGION \
		-e BUNNYNET_MC_ZONE \
		-e BUNNYNET_MC_APPID \
		-e BUNNYNET_MC_PODID \
		-e HACHYBOOP_TEST_FREQUENCY_SECONDS \
		--mount type=bind,src=data/,dst=/data \
	 	$(target):latest

.PHONY: install
install: ## Install the program to /usr/bin 🎉
	@echo "Installing..."
	sudo cp $(target) /usr/bin/$(target)

.PHONY: test
test: clean compile install ## 🤓 Run go tests
	@echo "Testing..."
	go test -v ./...

.PHONY: clean
clean: ## Clean your artifacts 🧼
	@echo "Cleaning..."
	rm -rvf release/*
	rm -rvf ./hachyboop

.PHONY: release
release: ## Make the binaries for a GitHub release 📦
	mkdir -p release
	GOOS="linux" GOARCH="amd64" go build -ldflags "-X 'github.com/$(org)/$(target).Version=$(version)'" -o release/$(target)-linux-amd64 cmd/*.go
	GOOS="linux" GOARCH="arm" go build -ldflags "-X 'github.com/$(org)/$(target).Version=$(version)'" -o release/$(target)-linux-arm cmd/*.go
	GOOS="linux" GOARCH="arm64" go build -ldflags "-X 'github.com/$(org)/$(target).Version=$(version)'" -o release/$(target)-linux-arm64 cmd/*.go
	GOOS="linux" GOARCH="386" go build -ldflags "-X 'github.com/$(org)/$(target).Version=$(version)'" -o release/$(target)-linux-386 cmd/*.go
	GOOS="darwin" GOARCH="amd64" go build -ldflags "-X 'github.com/$(org)/$(target).Version=$(version)'" -o release/$(target)-darwin-amd64 cmd/*.go

.PHONY: help
help:  ## 🤔 Show help messages for make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: querylast
querylast:
	duckdb -c "select * from read_parquet('${DATADIR}/$(shell ls -t data | head -n1)');"

.PHONY: queryall
queryall:
	duckdb -c "select * from read_parquet('data/*.parquet');"
