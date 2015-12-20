package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var (
	joinChan      = make(chan net.Conn)
	leftChan      = make(chan net.Conn)
	broadcastChan = make(chan string)
	conn2sender   = make(map[net.Conn]chan string)
)

func startMaster() {
	for {
		select {
		case c := <-joinChan:
			ch := make(chan string)
			conn2sender[c] = ch
			go sender(c, ch)

		case c := <-leftChan:
			delete(conn2sender, c)

		case msg := <-broadcastChan:
			for _, s := range conn2sender {
				go func(ch chan string) {
					ch <- msg
				}(s)
			}
		}
	}
}

func sender(c net.Conn, ch chan string) {
	for {
		msg := <-ch
		fmt.Fprintln(c, msg)
	}
}

func handleConn(c net.Conn) {
	name := fmt.Sprintf("Client %v", c.RemoteAddr().String())
	joinChan <- c
	broadcastChan <- fmt.Sprintf("%v joined", name)
	input := bufio.NewScanner(c)
	for input.Scan() {
		broadcastChan <- fmt.Sprintf("%v: %v", name, input.Text())
	}
	defer c.Close()
	leftChan <- c
	broadcastChan <- fmt.Sprintf("%v left", name)
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	go startMaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
