package main

import "io"

type ProgressReader struct {
	io.ReadCloser
	reader        io.ReadCloser
	total         int64
	onTotalUpdate func(int64)
	onProgress    func(int64)
}

func NewProgressReader(onTotalUpdate func(int64), onProgress func(int64)) ProgressReader {
	return ProgressReader{
		onTotalUpdate: onTotalUpdate,
		onProgress:    onProgress,
	}
}

func (pr ProgressReader) SetReader(reader io.ReadCloser) {
	pr.reader = reader
}

func (pr ProgressReader) SetTotal(n int64) {
	pr.total = n
	pr.onTotalUpdate(pr.total)
}

func (pr ProgressReader) Close() error {
	return pr.reader.Close()
}

func (pr ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.onProgress(int64(n))
	}
	return n, err
}
