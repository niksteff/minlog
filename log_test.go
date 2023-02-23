package minlog_test

import (
	"testing"

	"github.com/niksteff/minlog"
)

func BenchmarkDefault(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	logger := minlog.New()

	for i := 0; i < b.N; i++ {
		logger.Infof("this is run %d in a benchmark", i)
	}
}