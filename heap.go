package gohat

// Heap is a general structure for loading and analyzing a heap dump.
type Heap struct {
	Params           *DumpParams
	ITable           *ITable
	Objects          []*Object
	Goroutines       []*Goroutine
	OSThreads        []*Thread
	DataSegment      []*Segment
	BSS              []*Segment
	Finalizers       []*Finalizer
	QueuedFinalizers []*Finalizer
	MemStats         *MemStats
	MemProfs         []*MemProf

	curG *Goroutine
	curM *MemProf
}

// ITable contains types and itabs
type ITable struct {
	Types []*Type
	ITabs []*ITab
}

// NewHeap loads the heap dump in file and returns a Heap.
// The Add methods conform to the Heaper interface.
func NewHeap(file string) (*Heap, error) {
	h := &Heap{
		ITable: &ITable{},
	}

	err := Parse(file, h)
	return h, err
}

// AddParams adds the dump params to the Heap
func (h *Heap) AddParams(p *DumpParams) {
	h.Params = p
}

// AddType adds the type to the heap
func (h *Heap) AddType(t *Type) {
	h.ITable.Types = append(h.ITable.Types, t)
}

// AddITab adds an Itab entry to the heap
func (h *Heap) AddITab(i *ITab) {
	h.ITable.ITabs = append(h.ITable.ITabs, i)
}

// AddObject adds an object to the heap
func (h *Heap) AddObject(o *Object) {
	h.Objects = append(h.Objects, o)
}

// AddGoroutine adds a goroutine to the heap
func (h *Heap) AddGoroutine(g *Goroutine) {
	h.curG = g
	h.Goroutines = append(h.Goroutines, g)
}

// AddStackFrame associates the stack frame with the last added goroutine
func (h *Heap) AddStackFrame(f *StackFrame) {
	h.curG.StackFrames = append(h.curG.StackFrames, f)
}

// AddDefer associates the defer record to the last added goroutine
func (h *Heap) AddDefer(d *DeferRecord) {
	h.curG.DeferRecords = append(h.curG.DeferRecords, d)
}

// AddPanic associates the panic record to the last added goroutine
func (h *Heap) AddPanic(p *PanicRecord) {
	h.curG.PanicRecords = append(h.curG.PanicRecords)
}

// AddOSThread adds an os thread to the heap
func (h *Heap) AddOSThread(t *Thread) {
	h.OSThreads = append(h.OSThreads, t)
}

// AddDataSegment adds a datasegment to the heap
func (h *Heap) AddDataSegment(s *Segment) {
	h.DataSegment = append(h.DataSegment, s)
}

// AddBSS adds a bss segment to the heap
func (h *Heap) AddBSS(s *Segment) {
	h.BSS = append(h.BSS, s)
}

// AddFinalizer adds a finalizer to the heap
func (h *Heap) AddFinalizer(f *Finalizer) {
	h.Finalizers = append(h.Finalizers, f)
}

// AddQueuedFinalizer adds a queued finalizer to the heap
func (h *Heap) AddQueuedFinalizer(f *Finalizer) {
	h.QueuedFinalizers = append(h.QueuedFinalizers, f)
}

// AddMemStats adds the memstats to the heap
func (h *Heap) AddMemStats(m *MemStats) {
	h.MemStats = m
}

// AddMemProf adds a memprof entry to the heap
func (h *Heap) AddMemProf(m *MemProf) {
	h.curM = m
	h.MemProfs = append(h.MemProfs, m)
}

// AddAllocSample associates the sample with the last memprof added
func (h *Heap) AddAllocSample(s *AllocSample) {
	h.curM.Samples = append(h.curM.Samples, s)
}
