all: http webserver

OPTS =

http: http.go base64/base64.go
	go build -o http $(OPTS) http.go

webserver: webserver.go
	go build -o webserver $(OPTS) webserver.go

clean:
	rm -f http
	rm -f webserver

