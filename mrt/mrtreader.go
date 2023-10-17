package mrt

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/rspamd/goasn/log"

	"github.com/osrg/gobgp/v3/pkg/packet/mrt"
	"go.uber.org/zap"
)

type MRTReader struct {
	f       io.ReadCloser
	scanner *bufio.Scanner
	reject  io.WriteCloser
}

func NewMRTReader(fileName string, rejectName string) (*MRTReader, error) {
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

	if rejectName != "" {
		rdr.reject, err = os.Create(rejectName)
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
	if err != nil {
		log.Logger.Debug("mrt parsing failure", zap.Error(err))
		if m.reject != nil {
			m.reject.Write(b[:mrt.MRT_COMMON_HEADER_LEN+h.Len])
		}
	} else if m.reject != nil && message.Header.Type == mrt.TABLE_DUMPv2 && message.Header.SubType == uint16(mrt.PEER_INDEX_TABLE) {
		// always write peer index table to reject file
		m.reject.Write(b[:mrt.MRT_COMMON_HEADER_LEN+h.Len])
	}
	return more, message, err
}

func (m *MRTReader) Close() (err error) {
	if m.reject != nil {
		err = m.reject.Close()
	}
	return errors.Join(err, m.f.Close())
}
