HASH := $(shell git log -1 --pretty=format:"%H")
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s -X 'log-monitor.version=$(HASH)'" -a

verify:
	grep -a "log-monitor.version" log-monitor