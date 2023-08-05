# Makefile Copyright Alex Edwards, MIT licensed
# 
# https://www.alexedwards.net/blog/a-time-saving-makefile-for-your-go-projects
# https://gist.github.com/alexedwards/3b40775846535d0014ab1ff477e4a568


# HELPERS
# ========================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code > /dev/null
	git diff --staged --exit-code > /dev/null


# ========================================================================== #
# QUALITY CONTROL
# ========================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
# audit/vulncheck commented out until switch to 1.20.5 is complete.
audit: audit/lint audit/verify audit/format audit/vet audit/ineffassign audit/staticcheck audit/race

audit/verify:
	go mod verify

audit/format:
	test -z $$(gofmt -l .)

audit/vet:
	go vet ./...

audit/ineffassign:
	go run github.com/gordonklaus/ineffassign@latest ./...

audit/staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...

audit/vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

audit/race:
	go test -race -buildvcs -vet=off ./...

audit/lint:
	go run golang.org/x/lint/golint@latest -set_exit_status ./...


# ========================================================================== #
# DEVELOPMENT
# ========================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## build: build the application
.PHONY: build
build:
	# Include additional build steps, like TypeScript, SCSS or Tailwind compilation here...
	go build -v ./...


# ========================================================================== #
# OPERATIONS
# ========================================================================== #

## push: push changes to the remote Git repository
.PHONY: push
push: tidy audit no-dirty
	git push
