all: http webserver

OPTS =

http: http.go base64.c base64.h
	go build -o http $(OPTS) http.go base64.c

webserver: webserver.c
	gcc -o webserver $(OPTS) webserver.c

clean:
	rm -f http
	rm -f webserver
	rm -f core
	rm -f core.*
