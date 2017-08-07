package gohat

// These are the core types that are in the dump file. They are listed
// in the order WriteHeapDump writes them.

// DumpParams holds metadata about the heap dump
type DumpParams struct {
	BigEndian     bool
	PtrSize       uint64
	HeapStartAddr uint64
	HeapEndAddr   uint64
	Arch          string
	GoExperiment  string
	NCPU          uint64
}

// Type is a type
type Type struct {
	DescriptorAddr uint64
	TypeSize       uint64
	TypeName       string
	IsPtr          bool
}

// ITab is the itab
type ITab struct {
	TabPtr  uint64
	TypePtr uint64
}

// Object is an object
type Object struct {
	Addr   uint64
	Data   string
	Fields []Field
}

// Goroutine is a goroutine
type Goroutine struct {
	DescriptorAddr uint64
	StackPtr       uint64
	ID             uint64
	Creator        uint64
	Status         uint64
	IsSystem       bool
	IsBackground   bool
	LastWaitStart  uint64
	WaitReason     string
	FramePtr       uint64
	ThreadAddr     uint64
	TopDeferRecord uint64
	TopPanicRecord uint64
	StackFrames    []*StackFrame  `heap:"ignore"`
	DeferRecords   []*DeferRecord `heap:"ignore"`
	PanicRecords   []*PanicRecord `heap:"ignore"`
}

// StackFrame is a stack frame
type StackFrame struct {
	LastAddr uint64
	Depth    uint64
	ChildPtr uint64
	Contents string
	Entry    uint64
	PC       uint64
	ContinPC uint64
	FuncName string
	Fields   []Field
}

// DeferRecord is a defer record
type DeferRecord struct {
	Addr     uint64
	GAddr    uint64
	SP       uint64
	PC       uint64
	Val      uint64
	Fn       uint64
	NextAddr uint64
}

// PanicRecord is a panic record
type PanicRecord struct {
	Addr      uint64
	GAddr     uint64
	TypePtr   uint64
	DataField uint64
	X         uint64 // always 0
	NextAddr  uint64
}

// Thread is an os thread
type Thread struct {
	DescriptorAddr uint64
	GoID           uint64
	OSID           uint64
}

// Segment is a piece of data
type Segment struct {
	Addr     uint64
	Contents string
	Fields   []Field
}

// Finalizer is a finalizer
type Finalizer struct {
	ObjectAddr uint64
	FuncPtr    uint64
	PC         uint64
	ArgType    uint64
	ObjectType uint64
}

// MemStats is memstats
type MemStats struct {
	Alloc        uint64
	TotalAlloc   uint64
	Sys          uint64
	Lookups      uint64
	Mallocs      uint64
	Frees        uint64
	HeapAlloc    uint64
	HeapSys      uint64
	HeapIdle     uint64
	HeapInuse    uint64
	HeapReleased uint64
	HeapObjects  uint64
	StackInuse   uint64
	StackSys     uint64
	MSpanInuse   uint64
	MSpanSys     uint64
	MCacheInuse  uint64
	MCacheSys    uint64
	BuckHashSys  uint64
	GCSys        uint64
	OtherSys     uint64
	NextGC       uint64
	LastGC       uint64
	PauseTotalNs uint64
	PauseNs      [256]uint64
	NumGC        uint64
}

// MemProf is memprof
type MemProf struct {
	Addr      uint64
	Size      uint64
	NumFrames uint64
	FrameInfo []FrameInfo
	NumAllocs uint64
	NumFrees  uint64
	Samples   []*AllocSample `heap:"ignore"`
}

// FrameInfo is memprof frame info
type FrameInfo struct {
	Func string
	File string
	Line uint64
}

// AllocSample is an alloc sample
type AllocSample struct {
	Addr     uint64
	RecordID uint64
}

// Field is an object field
// This struct is loadable directly from the dump.
type Field struct {
	Kind   uint64
	Offset uint64
}
