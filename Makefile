 :all clean

all: build

build:
	go build -o celeve

clean:
	rm -f celeve

docker:
	docker build -t celeve --progress=plain .

docker-run:
	docker run -p 9898:9898 -p 23538:23538 celeve

docker-export: docker
	docker tag celeve 192.168.50.91:5000/celeve
	docker push 192.168.50.91:5000/celeve