default: build

build: dwcl

dwcl:
	@go install ./cmd/dwcl/...

