BINARY := bin/ak

.PHONY: build clean

build:
	GOOS=linux go build -o $(BINARY) .

clean:
	rm -f $(BINARY)
