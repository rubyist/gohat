package heapfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Alloc struct {
	objectAddress uint64 // address of object
	profileRecord uint64 // alloc/free profile record identifier
}

func (a *Alloc) Object() *Object {
	if obj, ok := objectList[a.objectAddress]; ok {
		return obj
	}
	return nil
}

func (a *Alloc) Profile() *Profile {
	if profile, ok := memProf[a.profileRecord]; ok {
		return profile
	}
	return nil
}

// Used for both data segment and bss
type Segment struct {
	Address uint64   // address of the start of the data segment
	Content string   // contents of the data segment
	Fields  []*Field // kind and offset of pointer-containing fields in the data segment.
}

// Returns objects the stack frame points to that are on the heap
func (s *Segment) Objects() []*Object {
	params := dumpParams
	var addr uint64
	var lastIndex uint64 = 0
	contentLength := uint64(len(s.Content))
	children := make([]*Object, 0)
	for i := params.PtrSize; i < contentLength+params.PtrSize; i += params.PtrSize {
		buf := bytes.NewReader([]byte(s.Content[lastIndex:i]))
		binary.Read(buf, binary.LittleEndian, &addr)
		lastIndex = i

		if obj, ok := objectList[addr]; ok {
			children = append(children, obj)
		}
	}
	return children
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

const (
	FieldPtr   = 1
	FieldStr   = 2
	FieldSlice = 3
	FieldIface = 4
	FieldEface = 5
)

type Field struct {
	Kind    uint64 // kind
	Offset  uint64 // offset
	Content string // contents of field
}

func (f *Field) String() string {
	return fmt.Sprintf("%s %d", f.KindString(), f.Offset)
}

func (f *Field) KindString() string {
	switch f.Kind {
	case FieldPtr:
		return "Ptr   "
	case FieldStr:
		return "String"
	case FieldSlice:
		return "Slice "
	case FieldIface:
		return "Iface "
	case FieldEface:
		return "Eface "
	}
	return ""
}

type Frame struct {
	Name string // function name
	File string // file name
	Line uint64 // line number
}

type Goroutine struct {
	Address       uint64 // address of descriptor
	Top           uint64 // pointer to the top of the stack (the currently running frame, a.k.a. depth 0)
	Id            uint64 // go routine ID
	Location      uint64 // the location of the go statement that created this routine
	status        uint64 // status
	System        bool   // is a Go routine started by the system
	Background    bool   // is a background Go routine
	LastWaiting   uint64 // approximate time the goroutine last started waiting (ns since epoc)
	reasonWaiting string // textual reason why it is waiting
	CurrentFrame  uint64 // context pointer of currently running frame
	OSThread      uint64 // address of os thread descriptor (M)
	DeferRecord   uint64 // top defer record
	PanicRecord   uint64 // top panic record
}

func (g *Goroutine) Status() string {
	switch g.status {
	case 0:
		return "idle"
	case 1:
		return "runnable"
	case 3:
		return "syscall"
	case 4:
		return "waiting"
	}
	return ""
}

func (g *Goroutine) ReasonWaiting() string {
	if g.status == 4 {
		return g.reasonWaiting
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

// Returns objects the object points to that are on the heap
func (o *Object) Children() []*Object {
	var lastIndex uint64 = 0
	var addr uint64
	var size uint64 = uint64(o.Size)
	children := make([]*Object, 0)

	if o.Size > 2252800 {
		return children
	}

	for i := dumpParams.PtrSize; i < size+dumpParams.PtrSize; i += dumpParams.PtrSize {
		buf := bytes.NewReader([]byte(o.Content[lastIndex:i]))
		binary.Read(buf, binary.LittleEndian, &addr)
		lastIndex = i

		if addr == o.Address {
			continue // Don't add ourselves
		}

		if child, ok := objectList[addr]; ok { // object is on the heap
			children = append(children, child)
		}
	}
	return children
}

type Finalizer struct {
	ObjectAddress uint64 // address of object that has a finalizer
	FuncValPtr    uint64 // pointer to FuncVal describing the finalizer
	PC            uint64 // PC of finalizer entry point
	ArgType       uint64 // type of finalizer argument
	ObjectType    uint64 // type of object
}

type Profile struct {
	Record    uint64 // record identifier
	Size      uint64 // size of allocated object
	NumFrames uint64 // number of stack frames
	Allocs    uint64 // number of allocations
	Frees     uint64 // number of frees
	Frames    []*Frame
}

type Root struct {
	Description string // textual description of where this root came from
	Pointer     uint64 // root pointer
}

type StackFrame struct {
	StackPointer      uint64   // stack pointer (lowest address inf rame)
	DepthInStack      uint64   // depth in stack (0 = top of stack)
	ChildFramePointer uint64   // stack pointer of child frame (or 0 if none)
	Content           string   // contents of stack frame
	EntryPC           uint64   // entry pc for function
	CurrentPC         uint64   // current pc for function
	ContinuationPC    uint64   // continuation pc for function (where functin may resume, if anywhere)
	Name              string   // function name
	FieldList         []*Field // list of kind and offset of pointer-containing fields in this frame
}

// Returns objects the stack frame points to that are on the heap
func (s *StackFrame) Objects() []*Object {
	params := dumpParams
	var addr uint64
	var lastIndex uint64 = 0

	contentLength := uint64(len(s.Content))
	children := make([]*Object, 0)

	for i := params.PtrSize; i < contentLength+params.PtrSize; i += params.PtrSize {
		buf := bytes.NewReader([]byte(s.Content[lastIndex:i]))
		binary.Read(buf, binary.LittleEndian, &addr)
		lastIndex = i

		if obj, ok := objectList[addr]; ok {
			children = append(children, obj)
		}
	}
	return children
}

type Type struct {
	Address   uint64   // address of type descriptor
	Size      uint64   // size of an object of thise type
	Name      string   // name of type
	IsPtr     bool     // whether the data field of an interface containing a value of this type is a pointer
	FieldList []*Field // a list of the kinds and locations of pointer-containing fields in objects of this type
}
