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

func (h *HeapFile) MemProf() []*Profile {
	h.parse()
	profiles := make([]*Profile, 0, len(memProf))
	for _, p := range memProf {
		profiles = append(profiles, p)
	}
	return profiles
}

func (h *HeapFile) Allocs() []*Alloc {
	h.parse()
	return allocs
}

func (h *HeapFile) Goroutines() []*Goroutine {
	h.parse()
	return goroutines
}

func (h *HeapFile) Roots() []*Root {
	h.parse()
	return roots
}

func (h *HeapFile) StackFrames() []*StackFrame {
	h.parse()
	return stackFrames
}
