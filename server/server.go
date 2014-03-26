package server

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var Statuses = map[int]string{
	400: "Bad Request",
	404: "Not Found",
	405: "Method Not Allowed",
	500: "Internal Server Error",
}

type Webs struct {
	Root string
	Addr string
	Port int
}

func New(root, addr string, port int) *Webs {
	return &Webs{root, addr, port}
}

func (webs *Webs) Listen() {
	laddr := fmt.Sprintf("%s:%d", webs.Addr, webs.Port)
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go webs.handleConnection(conn)
	}
}

func (webs *Webs) handleConnection(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("ERROR: ", err)
			httpError(500, conn)
		}
		conn.Close()
	}()


	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine()
	if err != nil {
		httpError(400, conn)
		return
	}

	tokens := strings.Split(string(line), " ")

	if len(tokens) != 3 {
		httpError(400, conn)
		return
	}

	if tokens[0] == "GET" {
		handleGet(conn, webs.Root, tokens[1])
	} else {
		httpError(405, conn)
	}
}

func handleGet(conn net.Conn, root, path string) {
	path, err := url.QueryUnescape(path)
	if err != nil {
		httpError(400, conn)
		return
	}
	abspath := filepath.Clean(filepath.Join(root, path))
	file, err := os.Open(abspath)
	if err != nil {
		httpError(404, conn)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	if stat.IsDir() {
		indexPath := filepath.Join(abspath, "index.html")
		idx, err := os.Open(indexPath)
		if err == nil {
			defer idx.Close()
			serveFile(conn, idx)
			return
		} else {
			serveDir(conn, file, path)
		}
	} else {
		serveFile(conn, file)
	}
}

func httpError(code int, conn net.Conn) {
	status, _ := Statuses[code]
	contentType := "Content-Type: text/html\r\n"
	server := "Server: Webs\r\n"

	fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n%s%s\r\n%s",
		code, status, contentType, server, status)
}

func serveDir(conn net.Conn, dir *os.File, path string) {
	files, _ := dir.Readdir(0)
	ctx := struct {
		Dir   string
		Files []os.FileInfo
	}{
		path,
		files,
	}
	fmt.Fprint(conn, "HTTP/1.1 200 ok\r\nServer: Webs\r\nContent-Type: text/html; charset=utf-8\r\n\r\n")
	tmpl.Execute(conn, ctx)
}

func serveFile(conn net.Conn, file *os.File) {
	t := mime.TypeByExtension(filepath.Ext(file.Name()))
	fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nServer: Webs\r\nContent-Type: %s\r\n\r\n", t)
	io.Copy(conn, file)
}

var tmpl = template.Must(template.New("dir").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Directory listing for{{.Dir}}</title>
</head>
<body>
	<h1>Directory listing for {{.Dir}}</h1>
	<hr>
	{{range .Files}}
		{{if .IsDir}}
			<li><a href="{{.Name}}/">{{.Name}}/</a></li>
		{{else}}
			<li><a href="{{.Name}}">{{.Name}}</a></li>
		{{end}}
	{{end}}
</body>
</html>`))
