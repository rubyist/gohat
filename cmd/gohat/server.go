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
	h := template.Must(template.New("header").Parse(headerTemplate))
	h.Execute(w, s.heapFile)

	t := template.Must(template.New("main").Parse(mainTemplate))
	t.Execute(w, s.heapFile.DumpParams())

	f := template.Must(template.New("footer").Parse(footerTemplate))
	f.Execute(w, nil)
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

var headerTemplate = `<html><head><title>GoHat - {{.Name}}</title></head><body>`
var footerTemplate = `</body></html>`

var mainTemplate = `
<h1>GoHat</h1>
<a href="/objects">All Objects</a>
<a href="/reachable">Reachable Objects</a>
<a href="/garbage">Garbage Objects</a>

<h2>Heap Parameters</h2>
<table>
<tr><td>Endianness</td><td>{{if .BigEndian}}Big{{else}}Little{{end}} Endian</td></tr>
<tr><td>Pointer Size</td><td>{{.PtrSize}}</td></tr>
<tr><td>Heap Start Address</td><td>{{.StartAddress}}</td></tr>
<tr><td>End Addres</td><td>{{.EndAddress}}</td></tr>
<tr><td>Arch</td><td>{{.Arch}}</td></tr>
<tr><td>GOEXPERIMENT</td><td>{{.GoExperiment}}</td></tr>
<tr><td>Num CPU</td><td>{{.NCPU}}</td></tr>
</table>
`
