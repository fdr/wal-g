package walg_test

import (
	"bytes"
	"github.com/wal-g/wal-g"
	"github.com/wal-g/wal-g/test_tools"
	"io"
	"io/ioutil"
	"testing"
)

func TestNoFilesProvided(t *testing.T) {
	buf := &tools.BufferTarInterpreter{}
	err := walg.ExtractAll(buf, []walg.ReaderMaker{})
	if err == nil {
		t.Errorf("extract: Did not catch no files provided error")
	}
}

func TestUnsupportedFileType(t *testing.T) {
	test := &bytes.Buffer{}
	brm := &BufferReaderMaker{test, "/usr/local", "gzip"}
	buf := &tools.BufferTarInterpreter{}
	files := []walg.ReaderMaker{brm}
	err := walg.ExtractAll(buf, files)

	if serr, ok := err.(*walg.UnsupportedFileTypeError); ok {
		t.Errorf("extract: Extract should not support filetype %s", brm.FileFormat)
	} else if serr != nil {
		t.Log(serr)
	}
}

// Tests roundtrip for a tar file.
func TestTar(t *testing.T) {
	//Generate and save random bytes compare against compression-decompression cycle.
	sb := tools.NewStrideByteReader(10)
	lr := &io.LimitedReader{
		R: sb,
		N: int64(1024),
	}
	b, err := ioutil.ReadAll(lr)

	//Copy generated bytes to another slice to make the test more robust against modifications of "b".
	bCopy := make([]byte, len(b))
	copy(bCopy, b)
	if err != nil {
		t.Fatal()
	}

	//Make a tar in memory.
	member := &bytes.Buffer{}
	tools.CreateTar(member, &io.LimitedReader{
		R: bytes.NewBuffer(b),
		N: int64(len(b)),
	})

	//Extract the generated tar and check that its one member is the same as the bytes generated to begin with.
	brm := &BufferReaderMaker{member, "/usr/local", "tar"}
	buf := &tools.BufferTarInterpreter{}
	files := []walg.ReaderMaker{brm}
	err = walg.ExtractAll(buf, files)
	if err != nil {
		t.Log(err)
	}

	if !bytes.Equal(bCopy, buf.Out) {
		t.Error("extract: Unbundled tar output does not match input.")
	}
}

// Used to mock files in memory.
type BufferReaderMaker struct {
	Buf        *bytes.Buffer
	Key        string
	FileFormat string
}

func (b *BufferReaderMaker) Reader() (io.ReadCloser, error) { return ioutil.NopCloser(b.Buf), nil }
func (b *BufferReaderMaker) Format() string                 { return b.FileFormat }
func (b *BufferReaderMaker) Path() string                   { return b.Key }
