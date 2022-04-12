package str

import (
	"bytes"
	"sync"
)

const (
	TooBigBlockSize = 1024 * 1024 * 4
)

var bytesBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func GetBytesBuffer() (data *bytes.Buffer) {
	data = bytesBufferPool.Get().(*bytes.Buffer)
	data.Reset()
	return data
}

func PutBytesBuffer(data *bytes.Buffer) {
	if data == nil || data.Len() > TooBigBlockSize {
		return
	}
	bytesBufferPool.Put(data)
}
