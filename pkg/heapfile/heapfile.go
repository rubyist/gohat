package heapfile

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

var (
	ErrInvalidHeapFile = errors.New("invalid heap file")
)

var dumpHeader = "go1.3 heap dump\n"

type HeapFile struct {
	Name       string
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

	name := filepath.Base(file)

	return &HeapFile{Name: name, byteReader: byteReader}, nil
}

func (h *HeapFile) DataSegment() *Segment {
	h.parse()
	return dataSegment
}

func (h *HeapFile) BSS() *Segment {
	h.parse()
	return bss
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

func (h *HeapFile) Garbage() []*Object {
	h.parse()
	objects := h.Objects()
	seen := make(map[uint64]bool, len(objects))

	for _, object := range objects {
		seen[object.Address] = false
	}

	// Mark all the objects the stack frames (roots) point to
	for _, frame := range h.StackFrames() {
		for _, object := range frame.Objects() {
			mark(object, &seen)
		}
	}

	// other roots
	for _, root := range h.Roots() {
		if object := h.Object(int64(root.Pointer)); object != nil {
			mark(object, &seen)
		}
	}

	// data segment
	for _, object := range h.DataSegment().Objects() {
		mark(object, &seen)
	}

	// bss
	for _, object := range h.BSS().Objects() {
		mark(object, &seen)
	}

	// finalizers
	for _, f := range h.QueuedFinalizers() {
		o := h.Object(int64(f.ObjectAddress))
		if o != nil {
			mark(o, &seen)

		}
	}
	for _, f := range h.Finalizers() {
		o := h.Object(int64(f.ObjectAddress))
		if o != nil {
			mark(o, &seen)

		}
	}

	trash := make([]*Object, 0, len(objects))
	for addr, visited := range seen {
		if !visited {
			trash = append(trash, h.Object(int64(addr)))
		}
	}

	return trash
}

func mark(object *Object, seen *map[uint64]bool) {
	if seen := (*seen)[object.Address]; seen {
		return
	}

	(*seen)[object.Address] = true
	for _, child := range object.Children() {
		mark(child, seen)
	}
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

func (h *HeapFile) QueuedFinalizers() []*Finalizer {
	h.parse()
	return queuedFinalizers
}

func (h *HeapFile) Finalizers() []*Finalizer {
	h.parse()
	return finalizers
}
