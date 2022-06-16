package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	host := "0.0.0.0"
	port := "9999"
	err := execute(host, port)
	if err != nil {
		os.Exit(1)
	}
}
func execute(host string, port string) (err error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(host, port)) // (запуск сервера длā прослушиваниā) мы ждем) когда клиент приходить обслуживаем
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() { // закрываем listener       закроем обслуживания
		cerr := listener.Close()
		if cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(err)
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		err = handler(conn)
		if err != nil {
			log.Print(err)
		}
	}
}
func handler(conn net.Conn) (err error) {
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(err)
		}
	}()

	// читать

	buff := make([]byte, 4096)
	n, err := conn.Read(buff)
	if err == io.EOF {
		log.Printf("%s", buff[:n])
		return nil
	}
	if err != nil {
		return err
	}

	data := buff[:n]
	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)
	if requestLineEnd == -1 {
		return
	}
	requestLine := string(data[:requestLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return
	}
	method, path, version := parts[0], parts[1], parts[2]
	if method != "GET" {
		return
	}
	if version != "HTTP/1.1" {
		return
	}
	if path == "/" {
		body, err := os.ReadFile("static/index.html")
		if err != nil {
			return fmt.Errorf("can't read index.html: %w", err)
		}
		marker := "{{year}}"
		year := time.Now().Year()
		body = bytes.ReplaceAll(body, []byte(marker), []byte(strconv.Itoa(year)))
		_, err = conn.Write([]byte(
			"HTTP/1.1 200 OK\r\n" +
				"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
				"Content-Type: text/html\r\n" +
				"Connection: close\r\n" +
				"\r\n" +
				string(body),
		))
		if err != nil {
			return err
		}
	}
	return nil
}
