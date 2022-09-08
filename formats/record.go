package formats

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"github.com/idertator/eogrec/models"
)

type RecordHeader struct {
	Count uint32
}

type Record struct {
	Filename string
	Header   RecordHeader
	pointer  *os.File
}

func CreateRecord(filename string) (*Record, error) {
	record := Record{
		Filename: filename,
		Header: RecordHeader{
			Count: 0,
		},
	}
	pointer, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	record.pointer = pointer

	err = record.WriteHeader()
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func ReadRecord(filename string) ([]models.Sample, error) {
	pointer, err := os.Open(filename)
	defer pointer.Close()

	if err != nil {
		return nil, err
	}

	header := RecordHeader{}

	raw := make([]byte, 4)
	_, err = pointer.Read(raw)
	if err != nil {
		return nil, err
	}

	buff := bytes.NewBuffer(raw)
	err = binary.Read(buff, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	rawSamples := make([]byte, 12*header.Count)
	_, err = pointer.Read(rawSamples)
	if err != nil {
		return nil, err
	}
	buffSamples := bytes.NewBuffer(rawSamples)

	samples := make([]models.Sample, header.Count)
	var i uint32

	for i = 0; i < header.Count; i++ {
		err = binary.Read(buffSamples, binary.BigEndian, &samples[i])
		if err != nil {
			return nil, err
		}
	}

	return samples, nil
}

func (r *Record) WriteHeader() error {
	offset, err := r.pointer.Seek(0, 0)
	if err != nil {
		return err
	}

	if offset != 0 {
		return errors.New("Cannot go to the start of the file")
	}

	var buff bytes.Buffer
	binary.Write(&buff, binary.BigEndian, r.Header)
	_, err = r.pointer.Write(buff.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (r *Record) AddSamples(samples []models.Sample, n uint32) error {
	var i uint32
	for i = 0; i < n; i++ {
		var buff bytes.Buffer
		binary.Write(&buff, binary.BigEndian, samples[i])
		_, err := r.pointer.Write(buff.Bytes())
		if err != nil {
			return err
		}
	}
	r.Header.Count += n
	return nil
}

func (r *Record) Close() error {
	err := r.WriteHeader()
	if err != nil {
		return err
	}
	return r.pointer.Close()
}
