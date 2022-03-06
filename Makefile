default: build

build: dwcl dwserver

dwcl:
	@go install ./cmd/dwcl/...

dwserver:
	@go install ./cmd/dwserver/...

container:
	docker build -f docker/Dockerfile -t bobbyho/dwserver:0.1 .

