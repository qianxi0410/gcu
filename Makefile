build:
	go build .

install:
	rm gcu && go install .

clean:
	rm gcu