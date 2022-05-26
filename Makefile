build:
	go build .

install: build
	go install .

clean: install
	rm gcu
