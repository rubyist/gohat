package heapfile

import (
	"bufio"
	"errors"
	"os"
	"runtime"
)

var (
	ErrInvalidHeapFile = errors.New("invalid heap file")
)

var dumpHeader = "go1.3 heap dump\n"

type HeapFile struct {
	memStats   *runtime.MemStats
	byteReader *bufio.Reader
	parsed     bool
}

func New(file string) (*HeapFile, error) {
	dumpFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	header := make([]byte, len(dumpHeader))
	dumpFile.Read(header)
	if string(header) != dumpHeader {
		return nil, ErrInvalidHeapFile
	}

	byteReader := bufio.NewReader(dumpFile)

	return &HeapFile{byteReader: byteReader}, nil
}

func (h *HeapFile) MemStats() *runtime.MemStats {
	h.parse()
	return h.memStats
}

func (h *HeapFile) Objects() []*Object {
	h.parse()
	objects := make([]*Object, 0, len(objectList))
	for _, v := range objectList {
		objects = append(objects, v)
	}
	return objects
}

func (h *HeapFile) Object(addr int64) *Object {
	h.parse()
	if object, ok := objectList[uint64(addr)]; ok {
		return object
	}
	return nil
}

func (h *HeapFile) Types() []*Type {
	h.parse()
	types := make([]*Type, 0, len(typeList))
	for _, t := range typeList {
		types = append(types, t)
	}
	return types
}

func (h *HeapFile) Type(addr int64) *Type {
	h.parse()
	return typeList[uint64(addr)]
}

func (h *HeapFile) DumpParams() *DumpParams {
	h.parse()
	return dumpParams
}

type Object struct {
	Address     uint64
	TypeAddress uint64
	Kind        uint64
	Content     string
	Size        int
	Type        *Type
}

type Type struct {
	Address   uint64
	Size      uint64
	Name      string
	IsPtr     uint64
	FieldList []Field
}

type Field struct {
	Kind   uint64
	Offset uint64
}

type DumpParams struct {
	BigEndian    bool   // big endian
	PtrSize      uint64 // pointer size in bytes
	ChHdrSize    uint64 // channel header size in bytes
	StartAddress uint64 // starting address of heap
	EndAddress   uint64 // ending address of heap
	Arch         uint64 // thechar = architecture specifier
	GoExperiment string // GOEXPERIMENT environment variable value
	NCPU         uint64 // runtime.ncpu
}
