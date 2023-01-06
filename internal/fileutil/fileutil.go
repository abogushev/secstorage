package fileutil

import (
	"bufio"
	"io"
	"os"
)

func Send(path string, chunkSender func([]byte) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	buffer := make([]byte, 4096)
	n := 0

	for {
		n, err = reader.Read(buffer)
		if err == io.EOF || n == 0 {
			return nil
		}
		if err != nil {
			return err
		}

		err = chunkSender(buffer[:n])
		if err != nil {
			return err
		}
	}
}

func Get(saveToPath string, chunkReceiver func() ([]byte, error)) error {
	file, err := os.Create(saveToPath)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	defer file.Close()
	defer writer.Flush()

	for {
		bytes, err := chunkReceiver()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		_, err = writer.Write(bytes)
		if err != nil {
			return err
		}
	}

}
