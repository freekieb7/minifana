package wal

import "os"

const (
	max_file_size  = 2e+7 // 20 MB
	max_frame_size = 65535
)

type Log struct {
	file os.File
}

//
//func (l *Log) Write(data []byte) (n int, err error) {
//	n := 0
//	dataLen := len(data)
//
//	for {
//		// Parts
//		data[n : n+max_frame_size]
//	}
//
//	for n < dl {
//		l.file.Write()
//		l.file.Write(data[n : n+8000])
//
//	}
//}
//
//func (l *Log) Read(format string, a ...interface{}) {
//	content, err := os.ReadFile("file.txt")
//}
//
//type fragment struct {
//	length [4]byte
//}
