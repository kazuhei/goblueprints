package trace

import (
	"fmt"
	"io"
)

// Tracer はコードの出来事を記録できるオブジェクトを表すインターフェースです
type Tracer interface {
	Trace(...interface{})
}

// New はtracerを作るコマンドです
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

type tracer struct {
	out io.Writer
}

func (t *tracer) Trace(a ...interface{}) {
	t.out.Write([]byte(fmt.Sprint(a...)))
	t.out.Write([]byte("\n"))
}

type nilTracer struct{}

func (t *nilTracer) Trace(a ...interface{}) {}

// Off は何もしないTracerを返します
func Off() Tracer {
	return &nilTracer{}
}
