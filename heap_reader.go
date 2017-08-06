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
	MemStats         *MemStats // can this just be a runtime.MemStats?
	MemProf          *MemProf
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

	heap := &Heap{}

	for parser.More() {
		tag := parser.Tag()

		obj, err := parser.Read()
		if err != nil {
			return nil, err
		}

		switch tag {
		case tagParams:
			heap.Params = obj.(*DumpParams)
		case tagObject:
			heap.Objects = append(heap.Objects, obj.(*Object))
		}
	}

	if err := parser.Error(); err != nil {
		return nil, err
	}

	return heap, nil
}

const dumpHeader = "go1.7 heap dump\n"

var (
	errInvalidHeapFile = errors.New("invalid heap file")
)
