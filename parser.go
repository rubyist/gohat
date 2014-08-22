package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime"
)

var dumpHeader = "go1.3 heap dump\n"
var typeList map[uint64]*Type
var objectList map[uint64]*Object

type HeapFile struct {
	File     string
	memStats *runtime.MemStats
}

func NewHeapFile(file string) (*HeapFile, error) {
	return &HeapFile{File: file}, nil
}

func (h *HeapFile) parse(contentObj uint64) {
	typeList = make(map[uint64]*Type, 0)
	objectList = make(map[uint64]*Object, 0)

	dumpFile, err := os.Open(h.File)
	if err != nil {
		fmt.Println("Error opening dump file", err)
		os.Exit(1)
	}

	// File header is
	header := make([]byte, len(dumpHeader))
	dumpFile.Read(header)
	if string(header) != dumpHeader {
		fmt.Println("Invalid dump file")
		os.Exit(1)
	}

	byteReader := bufio.NewReader(dumpFile)

	for {
		// From here on out is a series of records, starting with a uvarint
		kind, err := binary.ReadUvarint(byteReader)
		if err != nil {
			fmt.Println("Error reading:", err)
			os.Exit(1)
		}

		switch kind {
		case 0:
			return
		case 1:
			o := readObject(byteReader, contentObj)
			objectList[o.Address] = o
		case 2:
			readOtherRoot(byteReader)
		case 3:
			t := readType(byteReader)
			typeList[t.Address] = t
		case 4:
			readGoroutine(byteReader)
		case 5:
			readStackFrame(byteReader)
		case 6:
			readDumpParams(byteReader)
		case 7:
			readFinalizer(byteReader)
		case 8:
			readiTab(byteReader)
		case 9:
			readOSThread(byteReader)
		case 10:
			h.memStats = readMemStats(byteReader)
		case 11:
			readQueuedFinalizer(byteReader)
		case 12:
			readDataSegment(byteReader)
		case 13:
			readBSS(byteReader)
		case 14:
			readDeferRecord(byteReader)
		case 15:
			readPanicRecord(byteReader)
		case 16:
			readAllocFree(byteReader)
		case 17:
			readAllocSampleRecord(byteReader)
		default:
			fmt.Println("Unknown object kind")
			os.Exit(1)
		}
	}
}

func (h *HeapFile) Objects() []*Object {
	objects := make([]*Object, len(objectList))
	h.parse(0)
	for _, v := range objectList {
		objects = append(objects, v)
	}
	return objects
}

func (h *HeapFile) Object(addr int64) *Object {
	h.parse(uint64(addr))
	if object, ok := objectList[uint64(addr)]; ok {
		return object
	}
	return nil
}

func (h *HeapFile) MemStats() *runtime.MemStats {
	h.parse(0)
	return h.memStats
}

// (1) object: uvarint uvarint uvarint string
type Object struct {
	Address     uint64
	TypeAddress uint64
	Kind        uint64
	Content     string
	Size        int
	Type        *Type
}

func readObject(r io.ByteReader, contentObj uint64) *Object {
	o := &Object{}
	o.Address = readUvarint(r)     // address of object
	o.TypeAddress = readUvarint(r) // address of type descriptor (or 0 if unknown)
	o.Kind = readUvarint(r)        // kind of object (0=regular 1=array 2=channel 127=conservatively scanned)
	content := readString(r)       // contents of object (discard?)
	o.Size = len(content)
	if o.TypeAddress != 0 {
		o.Type = typeList[o.TypeAddress]
	}
	if contentObj == o.Address {
		o.Content = content
	}
	return o
}

// (2) other root
func readOtherRoot(r io.ByteReader) {
	readString(r)  // textual description of where this root came from
	readUvarint(r) // root pointer
}

// (3) type: uvarint uvarint string bool fieldlist
type Type struct {
	Address uint64
	Size    uint64
	Name    string
	IsPtr   uint64
	// FieldList
}

func readType(r io.ByteReader) *Type {
	t := &Type{}
	t.Address = readUvarint(r) // address of type descriptor
	t.Size = readUvarint(r)    // size of an object of this type
	t.Name = readString(r)     // name of type
	t.IsPtr = readUvarint(r)   // (bool) whether the data field of an interface containing a value of this type is a pointer
	readFieldList(r)           // a list of the kinds and locations of pointer-containing fields in objects of this type
	return t
}

// (4) goroutine
func readGoroutine(r io.ByteReader) {
	readUvarint(r) // address of descriptor
	readUvarint(r) // pointer to the top of the stack (the currently running frame, a.k.a. depth 0)
	readUvarint(r) // go routine ID
	readUvarint(r) // the locatio nof the go statement that created this routine
	readUvarint(r) // status
	readUvarint(r) // is a Go routine started by the system
	readUvarint(r) //is a background Go routine
	readUvarint(r) // approximate time the goroutine last started waiting (ns since epoc)
	readString(r)  // textual reason why it is waiting
	readUvarint(r) // context pointer of currently running frame
	readUvarint(r) // address of os thread descriptor (M)
	readUvarint(r) // top defer recort
	readUvarint(r) // top panic record
}

// (5) stackframe
func readStackFrame(r io.ByteReader) {
	readUvarint(r)   // stack pointer (lowest address in frame)
	readUvarint(r)   // depth in stack (0 = top of stack)
	readUvarint(r)   // stack pointer of child frame (or 0 if none)
	readString(r)    // contents of stack frame
	readUvarint(r)   // entry pc for function
	readUvarint(r)   // current pc for function
	readUvarint(r)   // continuation pc for function (where function may resume, if anywhere)
	readString(r)    // function name
	readFieldList(r) // list of kind and offset of pointer-containing fields in this frame
}

// (6) dump params: bool uvarint uvarint uvarint uvarint uvarint string varint
func readDumpParams(r io.ByteReader) {
	readUvarint(r) // (bool) big endian
	readUvarint(r) // pointer size in bytes
	readUvarint(r) // channel header size in bytes
	readUvarint(r) // starting address of heap
	readUvarint(r) // ending address of heap
	readUvarint(r) // thechar = architecture specifier
	readString(r)  // GOEXPERIMENT environment variable value
	readUvarint(r) // runtime.ncpu
}

// (7) registered finalizer
func readFinalizer(r io.ByteReader) {
	readUvarint(r) // address of object that has a finalizer
	readUvarint(r) // pointer to FuncVal describing the finalizer
	readUvarint(r) // PC of finalizer entry point
	readUvarint(r) // type of finalizer argument
	readUvarint(r) // type of object
}

// (8) itab: uvarint bool
func readiTab(r io.ByteReader) {
	readUvarint(r) // Itab address
	readUvarint(r) // (bool) whether the data field of an Iface with this itab is a pointer
}

// (9) os thread
func readOSThread(r io.ByteReader) {
	readUvarint(r) // address of this os thread descriptor
	readUvarint(r) // Go internal id of thread
	readUvarint(r) // os's id for thread
}

// (10) memstats
func readMemStats(r io.ByteReader) *runtime.MemStats {
	var memStats runtime.MemStats
	memStats.Alloc = readUvarint(r)        // bytes allocated and still in use
	memStats.TotalAlloc = readUvarint(r)   // bytes allocated (even if freed)
	memStats.Sys = readUvarint(r)          // bytes obtained from system (sum of XxxSys below)
	memStats.Lookups = readUvarint(r)      // number of pointer lookups
	memStats.Mallocs = readUvarint(r)      // number of mallocs
	memStats.Frees = readUvarint(r)        // number of frees
	memStats.HeapAlloc = readUvarint(r)    // bytes allocated and still in use
	memStats.HeapSys = readUvarint(r)      // bytes obtained from system
	memStats.HeapIdle = readUvarint(r)     // bytes in idle spans
	memStats.HeapInuse = readUvarint(r)    // bytes in non-idle span
	memStats.HeapReleased = readUvarint(r) // bytes released to the OS
	memStats.HeapObjects = readUvarint(r)  // total number of allocated objects
	memStats.StackInuse = readUvarint(r)   // bootstrap stacks
	memStats.StackSys = readUvarint(r)
	memStats.MSpanInuse = readUvarint(r) // mspan structures
	memStats.MSpanSys = readUvarint(r)
	memStats.MCacheInuse = readUvarint(r) // mcache structures
	memStats.MCacheSys = readUvarint(r)
	memStats.BuckHashSys = readUvarint(r) // profiling bucket hash table
	memStats.GCSys = readUvarint(r)       // GC metadata
	memStats.OtherSys = readUvarint(r)    // other system allocations
	memStats.NextGC = readUvarint(r)      // next run in HeapAlloc time (bytes)
	memStats.LastGC = readUvarint(r)      // last run in absolute time (ns)
	memStats.PauseTotalNs = readUvarint(r)
	for i := 0; i < 256; i++ {
		memStats.PauseNs[i] = readUvarint(r)
	}
	memStats.NumGC = uint32(readUvarint(r))
	return &memStats
}

// (11) queued finalizer
func readQueuedFinalizer(r io.ByteReader) {
	readUvarint(r) // address of object that has a finalizer
	readUvarint(r) // pointer to FuncVal describing the finalizer
	readUvarint(r) // PC of finalizer entry point
	readUvarint(r) // type of finalizer argument
	readUvarint(r) // type of object
}

// (12) data segment
func readDataSegment(r io.ByteReader) {
	readUvarint(r)   // address of the start of the data segment
	readString(r)    // contents of the data segment
	readFieldList(r) // kind and offset of pointer-containing fields in the data segment.
}

// (13) bss
func readBSS(r io.ByteReader) {
	readUvarint(r)   // address of the start of the data segment
	readString(r)    // contents of the data segment
	readFieldList(r) // kind and offset of pointer-containing fields in the data segment.
}

// (14) defer record
func readDeferRecord(r io.ByteReader) {
	readUvarint(r) // defer record address
	readUvarint(r) // containing goroutine
	readUvarint(r) // argp
	readUvarint(r) // pc
	readUvarint(r) // FuncVal of defer
	readUvarint(r) // PC of defer entry point
	readUvarint(r) // link to next defer record
}

// (15) panic record
func readPanicRecord(r io.ByteReader) {
	readUvarint(r) // panic record address
	readUvarint(r) // containing goroutine
	readUvarint(r) // type ptr of panic arg eface
	readUvarint(r) // data field of panic arg eface
	readUvarint(r) // ptr to defer record that's currently running
	readUvarint(r) // link to next panic record
}

// (16) alloc/free profile record
func readAllocFree(r io.ByteReader) {
	readUvarint(r)           // record identifier
	readUvarint(r)           // size of allocated object
	frames := readUvarint(r) // number of stack frames. For each frame:

	for i := 0; i < int(frames); i++ {
		readString(r)  // function name
		readString(r)  // file name
		readUvarint(r) // line number
	}

	readUvarint(r) // number of allocations
	readUvarint(r) // number of frees
}

// (17) alloc stack trace sample
func readAllocSampleRecord(r io.ByteReader) {
	readUvarint(r) // address of object
	readUvarint(r) // alloc/free profile record identifier
}

func readUvarint(r io.ByteReader) uint64 {
	v, err := binary.ReadUvarint(r)
	if err != nil {
		fmt.Println("Error reading:", err)
		os.Exit(1)
	}
	return v
}

func readString(r io.ByteReader) string {
	l := readUvarint(r)
	var by []byte
	for i := 0; i < int(l); i++ {
		b, err := r.ReadByte()
		if err != nil {
			fmt.Println("Error reading string:", err)
			os.Exit(1)
		}
		by = append(by, b)
	}
	return string(by)
}

type Field struct {
	Kind   uint64
	Offset uint64
}

func readFieldList(r io.ByteReader) []Field {
	fields := make([]Field, 0)
	var kind uint64
	var offset uint64

	kind = readUvarint(r)
	for kind != 0 {
		offset = readUvarint(r)
		fields = append(fields, Field{kind, offset})
		kind = readUvarint(r)
	}

	return fields
}
