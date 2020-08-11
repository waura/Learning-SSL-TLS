package main

import (
	"fmt"
	"os"
	"syscall"
	"errors"
)

const HTTP_PORT = 80
const DEFAULT_LINE_LEN = 255


func read_line(connection int) ([]byte, error) {
	var line []byte
	c := make([]byte, 1, 1)

	for {
		_, err := syscall.Read(connection, c)
		if err != nil {
			return nil, errors.New("failed to read")
			break
		}
    if (c[0] == '\n') && (line[len(line) - 1] == '\r') {
			return line[:len(line) - 1], nil
    }
 		line = append(line, c[0])
	}
  return line, nil;
}

func build_success_response(connection int) (error) {
  buf := "HTTP/1.1 200 Success\r\nConnection: Close\r\nContent-Type:text/html\r\n\r\n<html><head><title>Test Page</title></head><body>Nothing here</body></html>\r\n"

	// Technically, this should account for short writes.
	_, err := syscall.Write(connection, []byte(buf))
  if err != nil {
		return errors.New("Trying to respond")
	}
	return nil
}

func build_error_response(connection, error_code int) (error) {
	buf := "HTTP/1.1 " + string(error_code) + " Error Occurred\r\n\r\n"

	// Technically, this should account for short writes.
	_, err := syscall.Write(connection, []byte(buf))
  if err != nil {
		return errors.New("Trying to respond")
	}
	return nil
}

func process_http_request(connection int) {
	request_line, _ := read_line(connection)
	
	if string(request_line[:3]) != "GET" {
		// Only supports "GET" requests
    build_error_response(connection, 501)
  } else {
   // Skip over all header lines, don't care
    for {
			line, _ := read_line(connection)
			if string(line) == "" {
				break
			}
		}
    build_success_response(connection)
  }

	err := syscall.Close(connection)
  if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to close connection")
  }
}

func main() {
	
	listen_sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create listening socket: %s", err.Error())
		os.Exit(0)
	}

	on := 1
	err = syscall.SetsockoptInt(listen_sock, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, on)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Setting socket option: %s", err.Error())
		os.Exit(0)
	}

	var sa = &syscall.SockaddrInet4 {
		Port: HTTP_PORT,
		Addr: [4]byte {0, 0, 0, 0},
	}
  if err = syscall.Bind(listen_sock, sa); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to bind to local address")
    os.Exit(0)
	}
	
	err = syscall.Listen(listen_sock, syscall.SOMAXCONN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set socket backlog")
		os.Exit(0)
	}

	for {
		connect_sock, _, err := syscall.Accept(listen_sock)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to accept socket")
			break
		}
    // TODO: ideally, this would spawn a new thread.
    process_http_request(connect_sock)
	}

}