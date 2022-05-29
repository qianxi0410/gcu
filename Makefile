build:
	go build .

install:
	go install .

clean: build
	rm gcu
