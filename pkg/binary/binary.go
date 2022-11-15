package binary

import (
	"encoding/binary"
	"fmt"
	"io"
)

func WriteString(w io.Writer, s string) error {
	data := []byte(s)
	dataLen := uint32(len(data))
	err := binary.Write(w, binary.LittleEndian, dataLen)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func ReadString(r io.Reader) (string, error) {
	var dataLen uint32
	err := binary.Read(r, binary.LittleEndian, &dataLen)
	if err != nil {
		return "", err
	}
	data := make([]byte, int(dataLen))
	readLen, err := r.Read(data)
	if err != nil {
		return "", err
	}
	if readLen != int(dataLen) {
		return "", fmt.Errorf("expect to read %d data, but read %d", dataLen, readLen)
	}
	return string(data), nil
}

func WriteInt64(w io.Writer, i uint64) error {
	return binary.Write(w, binary.LittleEndian, i)
}

func ReadInt64(r io.Reader) (uint64, error) {
	var data uint64
	err := binary.Read(r, binary.LittleEndian, &data)
	return data, err
}
