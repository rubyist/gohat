package gohat

import "os"

type offsetReader struct {
	o int64
	f *os.File
}

func newOffsetReader(f *os.File) *offsetReader {
	return &offsetReader{
		f: f,
	}
}

func (o *offsetReader) Offset() int64 {
	return o.o
}

func (o *offsetReader) Read(b []byte) (int, error) {
	n, err := o.f.Read(b)
	o.o += int64(n)
	return n, err
}

func (o *offsetReader) ReadByte() (byte, error) {
	b := make([]byte, 1)
	_, err := o.Read(b)
	return b[0], err
}

func (o *offsetReader) Seek(offset int64) error {
	_, err := o.f.Seek(offset, 1)
	o.o += offset
	return err
}
