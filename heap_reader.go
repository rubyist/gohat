package gohat

import "errors"

type Heaper interface {
	AddParams(*DumpParams)
	AddType(*Type)
	AddITab(*ITab)
	AddObject(*Object)
	AddGoroutine(*Goroutine)
	AddStackFrame(*StackFrame)
	AddDefer(*DeferRecord)
	AddPanic(*PanicRecord)
	AddOSThread(*Thread)
	AddDataSegment(*Segment)
	AddBSS(*Segment)
	AddFinalizer(*Finalizer)
	AddQueuedFinalizer(*Finalizer)
	AddMemStats(*MemStats)
	AddMemProf(*MemProf)
	AddAllocSample(*AllocSample)
}

// ITable contains types and itabs
type ITable struct {
	Types []*Type
	ITabs []*ITab
}

// Heap represents all the data in the heap
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

func NewHeap(file string) (*Heap, error) {
	h := &Heap{
		ITable: &ITable{},
	}

	err := Parse(file, h)
	return h, err
}

func (h *Heap) AddParams(p *DumpParams) {
	h.Params = p
}

func (h *Heap) AddType(t *Type) {
	h.ITable.Types = append(h.ITable.Types, t)
}

func (h *Heap) AddITab(i *ITab) {
	h.ITable.ITabs = append(h.ITable.ITabs, i)
}

func (h *Heap) AddObject(o *Object) {
	h.Objects = append(h.Objects, o)
}

func (h *Heap) AddGoroutine(g *Goroutine) {
	h.curG = g
	h.Goroutines = append(h.Goroutines, g)
}

func (h *Heap) AddStackFrame(f *StackFrame) {
	h.curG.StackFrames = append(h.curG.StackFrames, f)
}

func (h *Heap) AddDefer(d *DeferRecord) {
	h.curG.DeferRecords = append(h.curG.DeferRecords, d)
}

func (h *Heap) AddPanic(p *PanicRecord) {
	h.curG.PanicRecords = append(h.curG.PanicRecords)
}

func (h *Heap) AddOSThread(t *Thread) {
	h.OSThreads = append(h.OSThreads, t)
}

func (h *Heap) AddDataSegment(s *Segment) {
	h.DataSegment = append(h.DataSegment, s)
}

func (h *Heap) AddBSS(s *Segment) {
	h.BSS = append(h.BSS, s)
}

func (h *Heap) AddFinalizer(f *Finalizer) {
	h.Finalizers = append(h.Finalizers, f)
}

func (h *Heap) AddQueuedFinalizer(f *Finalizer) {
	h.QueuedFinalizers = append(h.QueuedFinalizers, f)
}

func (h *Heap) AddMemStats(m *MemStats) {
	h.MemStats = m
}

func (h *Heap) AddMemProf(m *MemProf) {
	h.curM = m
	h.MemProfs = append(h.MemProfs, m)
}

func (h *Heap) AddAllocSample(s *AllocSample) {
	h.curM.Samples = append(h.curM.Samples, s)
}

// Parse creates a Heap for a heap dump file
func Parse(file string, h Heaper) error {
	parser, err := NewParser(file)
	if err != nil {
		return err
	}

	for parser.More() {
		tag := parser.Tag()

		obj, err := parser.Read()
		if err != nil {
			return err
		}

		if setter, ok := setters[tag]; ok {
			setter(h, obj)
		} else {
			return errInvalidTag
		}
	}

	if err := parser.Error(); err != nil {
		return err
	}

	return nil
}

var (
	errInvalidHeapFile = errors.New("invalid heap file")
	setters            = map[uint64]func(Heaper, interface{}){
		tagParams:          addParams,
		tagType:            addType,
		tagItab:            addITab,
		tagObject:          addObject,
		tagGoroutine:       addGoroutine,
		tagStackFrame:      addStackFrame,
		tagDefer:           addDefer,
		tagPanic:           addPanic,
		tagOSThread:        addThread,
		tagData:            addDataSegment,
		tagBSS:             addBSS,
		tagFinalizer:       addFinalizer,
		tagQueuedFinalizer: addQueuedFinalizer,
		tagMemStats:        addMemStats,
		tagMemProf:         addMemProf,
		tagAllocSample:     addAllocSample,
	}
)

func addParams(h Heaper, p interface{}) {
	h.AddParams(p.(*DumpParams))
}

func addType(h Heaper, p interface{}) {
	h.AddType(p.(*Type))
}

func addITab(h Heaper, p interface{}) {
	h.AddITab(p.(*ITab))
}

func addObject(h Heaper, p interface{}) {
	h.AddObject(p.(*Object))
}

func addGoroutine(h Heaper, p interface{}) {
	h.AddGoroutine(p.(*Goroutine))
}

func addStackFrame(h Heaper, p interface{}) {
	h.AddStackFrame(p.(*StackFrame))
}

func addDefer(h Heaper, p interface{}) {
	h.AddDefer(p.(*DeferRecord))
}

func addPanic(h Heaper, p interface{}) {
	h.AddPanic(p.(*PanicRecord))
}

func addThread(h Heaper, p interface{}) {
	h.AddOSThread(p.(*Thread))
}

func addDataSegment(h Heaper, p interface{}) {
	h.AddDataSegment(p.(*Segment))
}

func addBSS(h Heaper, p interface{}) {
	h.AddBSS(p.(*Segment))
}

func addFinalizer(h Heaper, p interface{}) {
	h.AddFinalizer(p.(*Finalizer))
}

func addQueuedFinalizer(h Heaper, p interface{}) {
	h.AddQueuedFinalizer(p.(*Finalizer))
}

func addMemStats(h Heaper, p interface{}) {
	h.AddMemStats(p.(*MemStats))
}

func addMemProf(h Heaper, p interface{}) {
	h.AddMemProf(p.(*MemProf))
}

func addAllocSample(h Heaper, p interface{}) {
	h.AddAllocSample(p.(*AllocSample))
}
