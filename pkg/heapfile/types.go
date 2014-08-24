package heapfile

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

type Type struct {
	Address   uint64  // address of type descriptor
	Size      uint64  // size of an object of thise type
	Name      string  // name of type
	IsPtr     bool    // whether the data field of an interface containing a value of this type is a pointer
	FieldList []Field // a list of the kinds and locations of pointer-containing fields in objects of this type
}
