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

type Field struct {
	kind   uint64 // kind
	Offset uint64 // offset
}

func (f *Field) Kind() string {
	switch f.kind {
	case 1:
		return "Ptr"
	case 2:
		return "String"
	case 3:
		return "Slice"
	case 4:
		return "Iface"
	case 5:
		return "Eface"
	}
	return ""
}

type Object struct {
	Address     uint64 // address of object
	TypeAddress uint64 // address of type descriptor (or 0 if unknown)
	kind        uint64 // kind of object  (0=regular 1=array 2=channel 127=conservatively scanned)
	Content     string // contents of object
	Size        int    // size of contents
	Type        *Type
}

func (o *Object) Kind() string {
	switch o.kind {
	case 0:
		return "regular"
	case 1:
		return "array"
	case 2:
		return "channel"
	case 127:
		return "conservatively scanned"
	}
	return ""
}

type Type struct {
	Address   uint64  // address of type descriptor
	Size      uint64  // size of an object of thise type
	Name      string  // name of type
	IsPtr     bool    // whether the data field of an interface containing a value of this type is a pointer
	FieldList []Field // a list of the kinds and locations of pointer-containing fields in objects of this type
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
