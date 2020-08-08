package main

import (
	"os"
	"flag"
	"syscall"
	"net"
	"fmt"
	"strings"
	"strconv"
	"errors"
	"./base64"
)

const HTTP_PORT = 80
const BUFFER_SIZE = 255

func parse_url(uri (*string)) (string, string, error) {
	pos1 := strings.Index(*uri, "//")
	if -1 == pos1 {		
		return "", "", errors.New("invalid url: not found //")
	}

	pos2 := strings.Index((*uri)[pos1 + 2:], "/")
	if -1 == pos2 {
		return (*uri)[pos1 + 2:], "/", nil
	}
	pos2 += pos1 + 2
	return (*uri)[pos1 + 2 : pos2], (*uri)[pos2:], nil
}

func parse_proxy_param(proxy_spec (*string)) (string, int, string, string, error) {
	if (*proxy_spec)[:7] != "http://" {
		return "", 0, "", "", errors.New("invalid proxy spec")
	}
	login_sep := strings.Index((*proxy_spec)[:], "@")

	colon_sep := 6
	var proxy_user string
	var proxy_password string
	if login_sep != -1 {
		colon_sep = strings.Index((*proxy_spec)[7:], ":")
		if colon_sep == -1 {
			return "", 0, "", "", errors.New("Expected password in " + (*proxy_spec))
		}
		proxy_user = (*proxy_spec)[7 : 7 + colon_sep]
		proxy_password = (*proxy_spec)[colon_sep + 8 : login_sep]
	}

	proxy_port := HTTP_PORT
	var proxy_host string

	trailer_sep := strings.Index((*proxy_spec)[login_sep + 1:], "/")
	if trailer_sep == -1 {
		proxy_host = (*proxy_spec)[login_sep + 1 :]
	} else {
		proxy_host = (*proxy_spec)[login_sep + 1 : trailer_sep]
	}

	colon_sep = strings.Index(proxy_host, ":")
	if colon_sep != -1 {
		proxy_port, _ = strconv.Atoi(proxy_host[colon_sep + 1:])
		proxy_host = proxy_host[:colon_sep]
	}

	return proxy_host, proxy_port, proxy_user, proxy_password, nil
}

/**
 * Format and send an HTTP get command. The return value will be 0
 * on success, -1 on failure, with errno set appropriately. The caller
 * must then retrieve the response.
 */
func http_get(connection int, path string, host string, proxy_host string, proxy_user string, proxy_password string) (error) {
	var get_command string

	if proxy_host != "" {
		get_command = "GET http://" + host + "/" + path + " HTTP/1.1\r\n"
	} else {
		get_command = "GET /" + host + " HTTP/1.1\r\n"
	}

	_, err := syscall.Write(connection, []byte(get_command))
	if err != nil {
		return err
	}

	get_command = "Host: " + host + "\r\n"
	_, err = syscall.Write(connection, []byte(get_command))
	if err != nil {
		return err
	}

	if proxy_user != "" {
		auth_string, err := base64.Encode([]byte(proxy_user + ":" + proxy_password))
		get_command = "Proxy-Authorization: BASIC " + string(auth_string) + "\r\n"
		_, err = syscall.Write(connection, []byte(get_command))
		if err != nil {
			return err
		}
	}

	get_command = "Connection: close\r\n\r\n"
	_, err = syscall.Write(connection, []byte(get_command))
	if err != nil {
		return err
	}

	return nil
}

/**
 * Receive all data available on a connection and dump it to stdout
 */
func display_result(connection int) {
	received := 0
	var err error
  var recv_buf [BUFFER_SIZE]byte

  for {
		received, err = syscall.Read(connection, recv_buf[:])
		if received <= 0 {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read response \n", err.Error())
		}
    // recv_buf[ received ] = byte('\0')
    fmt.Printf("%s", recv_buf)
  }
  fmt.Printf("\n")
}

func main() {
	var (
		p = flag.String("p", "", "proxy")
	)
	flag.Parse()
	args := flag.Args()

	if (len(args) < 1) {
		fmt.Fprintf(os.Stderr, "Usage: %s: [-p http://[username:password@]proxy-host:proxy-port] <URL>\n", os.Args[0])
		os.Exit(1)
	}

	proxy_host, proxy_port, proxy_user, proxy_password, err := parse_proxy_param(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error - malformed proxy parameter '%s'.\n", *p)
		os.Exit(2)
	}

	fmt.Printf("proxy_host, proxy_port, proxy_user, proxy_password = %s, %d, %s, %s\n", proxy_host, proxy_port, proxy_user, proxy_password)

	host, path, err := parse_url(&args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error - malformed URL '%s'.\n", os.Args[1])
		os.Exit(1)
	}

	fmt.Printf("Connecting to host '%s'\n", host)

	// Step 1: open a socket connection on http port with the destination host.
	client_connection, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create local socket\n")
		os.Exit(2)
	}

	var host_addr *net.IPAddr
	if proxy_host != "" {
	 	fmt.Fprintf(os.Stderr, "Connecting to host '%s'\n", proxy_host)
	 	host_addr, err = net.ResolveIPAddr("ip", proxy_host)
	} else {
		host_addr, err = net.ResolveIPAddr("ip", host)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in name resolution: %s \n", err.Error())
		os.Exit(3)
	}

	fmt.Printf("resolved ip address: %s\n", host_addr)

	var sa syscall.SockaddrInet4
	if proxy_host != "" {
	 	sa.Port = proxy_port
	} else {
		sa.Port = HTTP_PORT
	}
	copy(sa.Addr[:], host_addr.IP.To4())

	err = syscall.Connect(client_connection, &sa)
	if err != nil {		
		fmt.Fprintf(os.Stderr, "Unable to connect to host: %s \n", err.Error())
		os.Exit(4)
	}

	fmt.Printf("Retrieving document: '%s'\n", path)

	http_get(client_connection, path, host, proxy_host, proxy_user, proxy_password);

	display_result(client_connection)

	fmt.Printf("Shutting down.\n")
}