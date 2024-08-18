 :all clean

all: build

build:
	go build -o celeve

clean:
	rm -f celeve

docker:
	docker build -t celeve --progress=plain .

docker-run:
	docker run -p 8989:8989 -p 23538:23538 celeve

docker-export: docker
	docker save celeve -o celeve.tar

docker-import:
	docker load -i celeve.tar