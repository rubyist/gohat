package main

import (
	"fmt"
	"github.com/rubyist/gohat/pkg/heapfile"
	"html/template"
	"log"
	"net/http"
)

type gohatServer struct {
	heapFile *heapfile.HeapFile
	address  string
}

func newGohatServer(address string, heapFile *heapfile.HeapFile) *gohatServer {

	return &gohatServer{heapFile, address}
}

func (s *gohatServer) Run() {
	http.HandleFunc("/", s.mainPage)
	http.HandleFunc("/objects", s.objectsPage)
	http.HandleFunc("/reachable", s.reachablePage)
	http.HandleFunc("/garbage", s.garbagePage)

	log.Printf("Serving %s on %s", s.heapFile.Name, s.address)
	log.Fatal(http.ListenAndServe(s.address, nil))
}

func (s *gohatServer) mainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Printf("[404] %s", r.URL)
		http.NotFound(w, r)
		return
	}

	render(w, mainTemplate, s.heapFile)
	log.Printf("[200] %s", r.URL)
}

func (s *gohatServer) objectsPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "objects")
}

func (s *gohatServer) reachablePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "reachable")
}

func (s *gohatServer) garbagePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "garbage")
}

func render(w http.ResponseWriter, templateString string, data interface{}) {
	t := template.Must(template.New("main").Parse(bodyTemplate))
	t.New("body").Parse(templateString)
	t.Execute(w, data)
}

var bodyTemplate = `<html>
<head><title>GoHat {{.Name}}</title>
<body>
<h1>GoHat</h1>
<a href="/">Main</a>
<a href="/objects">All Objects</a>
<a href="/reachable">Reachable Objects</a>
<a href="/garbage">Garbage Objects</a>
{{template "body" .}}
</body>
</html>
`

var mainTemplate = `
<h2>Heap Parameters</h2>
<table>
<tr><td>Endianness</td><td>{{if .DumpParams.BigEndian}}Big{{else}}Little{{end}} Endian</td></tr>
<tr><td>Pointer Size</td><td>{{.DumpParams.PtrSize}}</td></tr>
<tr><td>Heap Start Address</td><td>{{printf "0x%x" .DumpParams.StartAddress}}</td></tr>
<tr><td>End Addres</td><td>{{printf "0x%x" .DumpParams.EndAddress}}</td></tr>
<tr><td>Arch</td><td>{{.DumpParams.Arch}}</td></tr>
<tr><td>GOEXPERIMENT</td><td>{{.DumpParams.GoExperiment}}</td></tr>
<tr><td>Num CPU</td><td>{{.DumpParams.NCPU}}</td></tr>
</table>
`
