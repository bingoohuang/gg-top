package main

import (
	"bytes"
	"compress/gzip"
	"io"
)

// Gzipd decompress gzip data.
func Gzipd(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	if r, err = gzip.NewReader(b); err != nil {
		return
	}

	var resB bytes.Buffer
	if _, err = resB.ReadFrom(r); err != nil {
		return
	}

	resData = resB.Bytes()
	return
}

// Gzip compress the data.
func Gzip(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	if _, err = gz.Write(data); err != nil {
		return
	}

	if err = gz.Close(); err != nil {
		return
	}

	compressedData = b.Bytes()
	return
}
