package gohat

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"reflect"
)

var (
	errInvalidType = errors.New("invalid type")
	errInvalidTag  = errors.New("invalid tag")
)

type HeapParser struct {
	file   string
	f      *offsetReader
	err    error
	curTag uint64
}

func NewParser(file string) (*HeapParser, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	b := newOffsetReader(f)

	// Parse header
	header := make([]byte, len(dumpHeader))
	if _, err := b.Read(header); err != nil {
		return nil, err
	}

	if string(header) != dumpHeader {
		return nil, errInvalidHeapFile
	}

	return &HeapParser{
		file: file,
		f:    b,
	}, nil
}

func (p *HeapParser) More() bool {
	tag, err := p.readNextTag()
	if err != nil {
		p.err = err
		return false
	}

	// TODO if you call More() without having Read, set an error and return false.
	p.curTag = tag
	return tag != tagEOF
}

func (p *HeapParser) Tag() uint64 {
	return p.curTag
}

func (p *HeapParser) Read() (interface{}, error) {
	var o interface{}

	switch p.curTag {
	case tagObject:
		o = new(Object)
	case tagOtherRoot: // Not used
	case tagType:
		o = new(Type)
	case tagGoroutine:
		o = new(Goroutine)
	case tagStackFrame:
		o = new(StackFrame)
	case tagParams:
		o = new(DumpParams)
	case tagFinalizer:
		o = new(Finalizer)
	case tagItab:
		o = new(ITab)
	case tagOSThread:
		o = new(Thread)
	case tagMemStats:
		o = new(MemStats)
	case tagQueuedFinalizer:
		o = &Finalizer{Kind: QueuedFinalizer}
	case tagData:
		o = &Segment{Kind: DataSegment}
	case tagBSS:
		o = &Segment{Kind: BSS}
	case tagDefer:
		o = new(DeferRecord)
	case tagPanic:
		o = new(PanicRecord)
	case tagMemProf:
		o = new(MemProf)
	case tagAllocSample:
		o = new(AllocSample)
	default:
		return nil, nil
	}

	err := p.readInto(o)
	return o, err
}

func (p *HeapParser) Error() error {
	return p.err
}

func (p *HeapParser) readNextTag() (uint64, error) {
	t, err := readUvarint(p.f)
	if err != nil {
		return 0, err
	}
	if t > tagAllocSample {
		return 0, errInvalidTag
	}
	return t, nil
}

func (p *HeapParser) readInto(c interface{}) error {
	// Verify that type of s matches current tag
	if reflect.ValueOf(c).Kind() != reflect.Ptr {
		return errInvalidType
	}

	s := reflect.ValueOf(c).Elem()

	for i := 0; i < s.NumField(); i++ {
		tag, _ := s.Type().Field(i).Tag.Lookup("heap")
		if tag == "ignore" {
			continue
		}

		field := s.Field(i)
		switch field.Kind() {
		case reflect.Uint64:
			v, err := readUvarint(p.f)
			if err != nil {
				return err
			}
			field.SetUint(v)
		case reflect.Bool:
			v, err := readBool(p.f)
			if err != nil {
				return err
			}
			field.SetBool(v)
		case reflect.String:
			v, err := readString(p.f)
			if err != nil {
				return err
			}
			field.SetString(v)
		case reflect.Array:
			// The PauseNs array in MemStats
			v, err := readPauseNs(p.f)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(v))
		case reflect.Slice:
			// I'm sure there's a better way to do this
			switch field.Type().Elem() {
			case reflect.TypeOf(Field{}):
				v, err := readFields(p.f)
				if err != nil {
					return err
				}
				field.Set(reflect.ValueOf(v))
			case reflect.TypeOf(FrameInfo{}):
				// If this is the case, c must be a MemProf
				// and the size field will have been set previously
				m, ok := c.(*MemProf)
				if !ok {
					return errInvalidType
				}

				v, err := p.readFrameInfo(m.NumFrames)
				if err != nil {
					return err
				}

				field.Set(reflect.ValueOf(v))
			default:
				return errInvalidType
			}
		default:
			return errInvalidType
		}
	}

	return nil
}

func readUvarint(r io.ByteReader) (uint64, error) {
	v, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func readString(r *offsetReader) (string, error) {
	l, err := readUvarint(r)
	if err != nil {
		return "", err
	}
	by := make([]byte, l)
	_, err = r.Read(by)
	return string(by), err
}

func readBool(r io.ByteReader) (bool, error) {
	i, err := readUvarint(r)
	if err != nil {
		return false, err
	}
	return i == 1, nil
}

func readFields(r io.ByteReader) ([]Field, error) {
	fields := make([]Field, 0)

	for {
		v, err := readUvarint(r)
		if err != nil {
			return nil, err
		}

		if v == fieldKindEol {
			break
		}

		o, err := readUvarint(r)
		if err != nil {
			return nil, err
		}

		fields = append(fields, Field{Kind: v, Offset: o})
	}

	return fields, nil
}

func (p *HeapParser) readFrameInfo(n uint64) ([]FrameInfo, error) {
	frames := make([]FrameInfo, 0)

	for i := uint64(0); i < n; i++ {
		f := FrameInfo{}
		if err := p.readInto(&f); err != nil {
			return nil, err
		}

		frames = append(frames, f)
	}

	return frames, nil
}

func readPauseNs(r io.ByteReader) ([256]uint64, error) {
	f := [256]uint64{}
	for i := 0; i < 256; i++ {
		v, err := readUvarint(r)
		if err != nil {
			return f, err
		}
		f[i] = v
	}

	return f, nil
}

const (
	fieldKindEol   = 0
	fieldKindPtr   = 1
	fieldKindIface = 2
	fieldKindEface = 3

	tagEOF             = 0
	tagObject          = 1
	tagOtherRoot       = 2 // not used
	tagType            = 3
	tagGoroutine       = 4
	tagStackFrame      = 5
	tagParams          = 6
	tagFinalizer       = 7
	tagItab            = 8
	tagOSThread        = 9
	tagMemStats        = 10
	tagQueuedFinalizer = 11
	tagData            = 12
	tagBSS             = 13
	tagDefer           = 14
	tagPanic           = 15
	tagMemProf         = 16
	tagAllocSample     = 17
)
