package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/norahiko/webs/server"
)

func main() {
	var host string
	var port int

	flag.StringVar(&host, "h", "localhost", "Host name")
	flag.IntVar(&port, "p", 8000, "Port number")
	flag.Parse()
	root, _ := os.Getwd()

	webs := server.New(root, host, port)
	fmt.Printf("listening on %s:%d\n", host, port)
	webs.Listen()
}
