package safe_socket

import (
	"fmt"
	"io"
)

//TODO: Complete with a short-read/short-write tolerant implementation

func SendAll(socket io.Writer, bytes []byte) error {
	bytes_sent := 0
	size := len(bytes)

	for bytes_sent < size {
		n, err := socket.Write(bytes[bytes_sent:])
		bytes_sent += n

		if err != nil {
			return err
		}
	}

	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	bytes_read := 0

	for bytes_read < size {
		n, err := socket.Read(buff[bytes_read:])
		bytes_read += n

		if err != nil {
			if err == io.EOF && bytes_read < size {
				return nil, fmt.Errorf("conexión cerrada antes de leer los %d bytes requeridos (se leyeron %d)", size, bytes_read)
			}
			return nil, err
		}

	}

	return buff[:bytes_read], nil
}
