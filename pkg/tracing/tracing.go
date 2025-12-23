package tracing

import (
	"errors"
	"os"
	"path/filepath"
	"runtime/trace"
	"strings"
)

type Tracer struct {
	f        *os.File
	FileName string
}

func StartTracing(projectDir string) (*Tracer, error) {
	dir, ok := os.LookupEnv("PWD")
	if !ok {
		return nil, errors.New("env PWD not found")
	}

	// PWD contains go.mod directory, if run in IDE
	if !strings.HasSuffix(dir, projectDir) {
		dir = filepath.Join(dir, projectDir)
	}

	traceFileName := filepath.Join(dir, "trace")

	tr := &Tracer{
		FileName: traceFileName,
	}

	var err error

	tr.f, err = os.OpenFile(traceFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}

	err = trace.Start(tr.f)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

func (t *Tracer) Close() error {
	trace.Stop()

	return t.f.Close()
}
