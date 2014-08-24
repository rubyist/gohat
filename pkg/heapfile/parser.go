package heapfile

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime"
)

var typeList map[uint64]*Type
var objectList map[uint64]*Object
var dumpParams *DumpParams
var memProf map[uint64]*Profile
var allocs []*Alloc
var goroutines []*Goroutine

func (h *HeapFile) parse() {
	if h.parsed {
		return
	}

	typeList = make(map[uint64]*Type, 0)
	objectList = make(map[uint64]*Object, 0)
	memProf = make(map[uint64]*Profile, 0)
	allocs = make([]*Alloc, 0)
	goroutines = make([]*Goroutine, 0)

	for {
		// From here on out is a series of records, starting with a uvarint
		kind, err := binary.ReadUvarint(h.byteReader)
		if err != nil {
			fmt.Println("Error reading:", err)
			os.Exit(1)
		}

		switch kind {
		case 0:
			h.parsed = true
			return
		case 1:
			o := readObject(h.byteReader)
			objectList[o.Address] = o
		case 2:
			readOtherRoot(h.byteReader)
		case 3:
			t := readType(h.byteReader)
			typeList[t.Address] = t
		case 4:
			goroutines = append(goroutines, readGoroutine(h.byteReader))
		case 5:
			readStackFrame(h.byteReader)
		case 6:
			dumpParams = readDumpParams(h.byteReader)
		case 7:
			readFinalizer(h.byteReader)
		case 8:
			readiTab(h.byteReader)
		case 9:
			readOSThread(h.byteReader)
		case 10:
			h.memStats = readMemStats(h.byteReader)
		case 11:
			readQueuedFinalizer(h.byteReader)
		case 12:
			readDataSegment(h.byteReader)
		case 13:
			readBSS(h.byteReader)
		case 14:
			readDeferRecord(h.byteReader)
		case 15:
			readPanicRecord(h.byteReader)
		case 16:
			profile := readAllocFree(h.byteReader)
			memProf[profile.Record] = profile
		case 17:
			allocs = append(allocs, readAllocSampleRecord(h.byteReader))
		default:
			fmt.Println("Unknown object kind")
			os.Exit(1)
		}
	}
}

// (1) object: uvarint uvarint uvarint string
func readObject(r io.ByteReader) *Object {
	o := &Object{}
	o.Address = readUvarint(r)
	o.TypeAddress = readUvarint(r)
	o.kind = readUvarint(r)
	o.Content = readString(r)
	o.Size = len(o.Content)
	if o.TypeAddress != 0 {
		o.Type = typeList[o.TypeAddress]
	}
	return o
}

// (2) other root
func readOtherRoot(r io.ByteReader) {
	readString(r)  // textual description of where this root came from
	readUvarint(r) // root pointer
}

// (3) type: uvarint uvarint string bool fieldlist
func readType(r io.ByteReader) *Type {
	t := &Type{}
	t.Address = readUvarint(r)
	t.Size = readUvarint(r)
	t.Name = readString(r)
	t.IsPtr = readUvarint(r) == 1
	t.FieldList = readFieldList(r)
	return t
}

// (4) goroutine
func readGoroutine(r io.ByteReader) *Goroutine {
	g := &Goroutine{}
	g.Address = readUvarint(r)
	g.Top = readUvarint(r)
	g.Id = readUvarint(r)
	g.Location = readUvarint(r)
	g.status = readUvarint(r)
	g.System = readUvarint(r) == 1
	g.Background = readUvarint(r) == 1
	g.LastWaiting = readUvarint(r)
	g.reasonWaiting = readString(r)
	g.CurrentFrame = readUvarint(r)
	g.OSThread = readUvarint(r)
	g.DeferRecord = readUvarint(r)
	g.PanicRecord = readUvarint(r)
	return g
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
func readDumpParams(r io.ByteReader) *DumpParams {
	dumpParams := &DumpParams{}
	dumpParams.BigEndian = (readUvarint(r) == 0)
	dumpParams.PtrSize = readUvarint(r)
	dumpParams.ChHdrSize = readUvarint(r)
	dumpParams.StartAddress = readUvarint(r)
	dumpParams.EndAddress = readUvarint(r)
	dumpParams.Arch = readUvarint(r)
	dumpParams.GoExperiment = readString(r)
	dumpParams.NCPU = readUvarint(r)

	return dumpParams
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
func readAllocFree(r io.ByteReader) *Profile {
	profile := &Profile{}

	profile.Record = readUvarint(r)
	profile.Size = readUvarint(r)
	profile.NumFrames = readUvarint(r)

	profile.Frames = make([]*Frame, 0, profile.NumFrames)

	for i := 0; i < int(profile.NumFrames); i++ {
		frame := &Frame{}
		frame.Name = readString(r)
		frame.File = readString(r)
		frame.Line = readUvarint(r)
		profile.Frames = append(profile.Frames, frame)
	}

	profile.Allocs = readUvarint(r)
	profile.Frees = readUvarint(r)
	return profile
}

// (17) alloc stack trace sample
func readAllocSampleRecord(r io.ByteReader) *Alloc {
	alloc := &Alloc{}
	alloc.objectAddress = readUvarint(r)
	alloc.profileRecord = readUvarint(r)
	return alloc
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
