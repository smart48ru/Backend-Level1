package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	fmt.Println("Enter your nickname:")
	conn, err := net.Dial("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	go func() {
		_, err = io.Copy(os.Stdout, conn)
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = io.Copy(conn, os.Stdin) // until you send ^Z
	if err != nil {
		return
	}

	fmt.Printf("%s: exit", conn.LocalAddr())
}
