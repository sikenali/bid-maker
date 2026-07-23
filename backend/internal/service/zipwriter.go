package service

import (
	"archive/zip"
	"bytes"
	"io"
)

type ZipWriter struct {
	buf *bytes.Buffer
	w   *zip.Writer
}

func NewZipWriter(buf *bytes.Buffer) *ZipWriter {
	return &ZipWriter{
		buf: buf,
		w:   zip.NewWriter(buf),
	}
}

func (zw *ZipWriter) AddFile(name string, data []byte) error {
	f, err := zw.w.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(data))
	return err
}

func (zw *ZipWriter) Close() error {
	return zw.w.Close()
}