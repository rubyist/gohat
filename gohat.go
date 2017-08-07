package gohat

import "errors"

var (
	errInvalidHeapFile = errors.New("invalid heap file")
)

// Heaper is a beefy interface that works with the gohat heap
// parser. As things are parsed out of a heap dump file, these methods
// are called, allowing an object to build up its own representation
// of a heap.
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

// Parse parses a heap dump file, passing the parts along to a Heaper.
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

var setters = map[uint64]func(Heaper, interface{}){
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
