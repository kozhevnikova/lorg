package lorg

import (
	"bytes"
	"testing"
)

func BenchmarkLog_Printf_Parallel(b *testing.B) {
	const logString = "lorg"
	var buffer bytes.Buffer

	log := NewLog()
	log.SetFormat(NewFormat("${level:%s:right} %s"))
	log.SetOutput(&buffer)

	b.RunParallel(func(pb *testing.PB) {
		buffer.Reset()
		log.format.Reset()
		for pb.Next() {
			log.Printf("%v", logString)
		}
	})
}

func BenchmarkLog_Printf(b *testing.B) {
	const logString = "lorg"
	var buffer bytes.Buffer

	log := NewLog()
	log.SetOutput(&buffer)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		log.Printf("%v", logString)
	}
}