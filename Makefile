build:
	docker build -t sfc .

run:
	docker run -v $(PWD)/test/auth.yml:/secrets/auth.yml -it --rm -p 8080:8080  -e PORT=8080 sfc

all:
	make build
	make run