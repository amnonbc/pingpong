package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

var (
	port     = 8765
	msgLen   = 100
	nMessage = 10
	delay    = time.Second
	fork     = false
	pong     = false
)

func main() {
	flag.IntVar(&port, "port", 8765, "port")
	flag.IntVar(&msgLen, "len", 100, "message len")
	flag.IntVar(&nMessage, "num", 10, "number of messages to send")
	flag.DurationVar(&delay, "delay", time.Second, "delay between messages")
	flag.BoolVar(&fork, "fork", false, "run a separate process")
	flag.BoolVar(&pong, "pong", false, "just run server")
	flag.Parse()

	if pong {
		doEcho()
		return
	}
	if fork {
		doFork()
	} else {
		go doEcho()
	}
	doPing()
}

func doFork() {
	args := make([]string, len(os.Args))
	copy(args, os.Args)
	args = append(args, "-pong")
	cmd := exec.Command(args[0], args[1:]...)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

}

func ping() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := make([]byte, msgLen)
	tick := time.NewTicker(delay)
	defer tick.Stop()
	n := 0
	for range tick.C {
		n++
		if n >= nMessage {
			break
		}
		start := time.Now()
		_, err := conn.Write(msg)
		if err != nil {
			return err
		}
		_, err = conn.Read(msg)
		if err != nil {
			return err
		}
		fmt.Println("got packet", n, "in", time.Since(start))
	}
	copy(msg, "end")
	_, err = conn.Write(msg)
	return err
}

func doEcho() {
	err := echo()
	if err != nil {
		log.Println(err)
	}
}

func doPing() {
	err := ping()
	if err != nil {
		log.Println(err)
	}
}

func echo() error {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	msg := make([]byte, msgLen)
	for {
		_, err := io.ReadAtLeast(conn, msg, msgLen)
		if err != nil {
			log.Println(err)
			return
		}
		if bytes.HasPrefix([]byte("end"), msg) {
			return
		}
		_, err = conn.Write(msg)
		if err != nil {
			log.Println(err)
			return
		}
	}

}
