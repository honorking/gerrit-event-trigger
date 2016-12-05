build:
	go build -o gerrit-observatory main.go

pack:
	tar zcvf gerrit-observatory.tar.gz gerrit-observatory Makefile start.sh stop.sh .id_rsa

clean:
	rm -rf ./gerrit-observatory

start:
	sh start.sh

stop:
	sh stop.sh

restart:
	sh stop.sh
	sh start.sh
