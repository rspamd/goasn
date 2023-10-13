package mrt

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/osrg/gobgp/v3/pkg/packet/mrt"
)

type MRTReader struct {
	f       io.ReadCloser
	scanner *bufio.Scanner
}

func NewMRTReader(fileName string) (*MRTReader, error) {
	rdr := &MRTReader{}
	var err error

	switch filepath.Ext(fileName) {
	case ".gz":
		unc, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}
		rdr.f, err = gzip.NewReader(unc)
		if err != nil {
			return nil, err
		}
	default:
		rdr.f, err = os.Open(fileName)
		if err != nil {
			return nil, err
		}
	}

	rdr.scanner = bufio.NewScanner(rdr.f)
	rdr.scanner.Split(mrt.SplitMrt)
	return rdr, nil
}

func (m *MRTReader) Next() (bool, *mrt.MRTMessage, error) {
	more := m.scanner.Scan()
	b := m.scanner.Bytes()

	if len(b) == 0 && !more {
		return more, nil, nil
	}

	h := &mrt.MRTHeader{}
	err := h.DecodeFromBytes(b)
	if err != nil {
		return more, nil, err
	}
	message, err := mrt.ParseMRTBody(h, b[mrt.MRT_COMMON_HEADER_LEN:])
	return more, message, err
}

func (m *MRTReader) Close() error {
	return m.f.Close()
}
