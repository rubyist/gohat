package gohat

import "errors"

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
}

// ITable contains types and itabs
type ITable struct {
	Types []*Type
	ITabs []*ITab
}

// NewHeap creates a Heap for a heap dump file
func NewHeap(file string) (*Heap, error) {
	parser, err := NewParser(file)
	if err != nil {
		return nil, err
	}

	heap := &Heap{
		ITable: &ITable{},
	}

	var curG *Goroutine
	var curM *MemProf

	for parser.More() {
		tag := parser.Tag()

		obj, err := parser.Read()
		if err != nil {
			return nil, err
		}

		switch tag {
		case tagParams:
			heap.Params = obj.(*DumpParams)
		case tagType:
			t := obj.(*Type)
			heap.ITable.Types = append(heap.ITable.Types, t)
		case tagItab:
			i := obj.(*ITab)
			heap.ITable.ITabs = append(heap.ITable.ITabs, i)
		case tagObject:
			heap.Objects = append(heap.Objects, obj.(*Object))
		case tagGoroutine:
			g := obj.(*Goroutine)
			curG = g
			heap.Goroutines = append(heap.Goroutines, g)
		case tagStackFrame:
			f := obj.(*StackFrame)
			curG.StackFrames = append(curG.StackFrames, f)
		case tagDefer:
			d := obj.(*DeferRecord)
			curG.DeferRecords = append(curG.DeferRecords, d)
		case tagPanic:
			p := obj.(*PanicRecord)
			curG.PanicRecords = append(curG.PanicRecords, p)
		case tagOSThread:
			t := obj.(*Thread)
			heap.OSThreads = append(heap.OSThreads, t)
		case tagData:
			d := obj.(*Segment)
			heap.DataSegment = append(heap.DataSegment, d)
		case tagBSS:
			b := obj.(*Segment)
			heap.BSS = append(heap.BSS, b)
		case tagFinalizer:
			f := obj.(*Finalizer)
			heap.Finalizers = append(heap.Finalizers, f)
		case tagQueuedFinalizer:
			f := obj.(*Finalizer)
			heap.QueuedFinalizers = append(heap.QueuedFinalizers, f)
		case tagMemStats:
			s := obj.(*MemStats)
			heap.MemStats = s
		case tagMemProf:
			m := obj.(*MemProf)
			curM = m
			heap.MemProfs = append(heap.MemProfs, m)
		case tagAllocSample:
			s := obj.(*AllocSample)
			curM.Samples = append(curM.Samples, s)
		}
	}

	if err := parser.Error(); err != nil {
		return nil, err
	}

	return heap, nil
}

var (
	errInvalidHeapFile = errors.New("invalid heap file")
)
