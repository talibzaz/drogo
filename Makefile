prod:
	make dep
	make build-binary
	make build
	make remove-container
	make run-prod

dev:
	make build-binary
	make build
	make remove-container
	make run-dev

build-binary:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o drogo ./src/code.drogo.gw.com

build:
	docker build -t drogo .

run-dev:
	docker run \
	-p 8085:80 \
	--rm \
	--name drogo \
	drogo

run-prod:
	docker run -d \
	-p 80:80 \
	--rm \
	--name drogo \
	drogo

remove-container:
	docker rm --force drogo || true

dep:
	cd src/code.drogo.gw.com/gql && gqlgen