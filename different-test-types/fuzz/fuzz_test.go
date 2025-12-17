package fuzz

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func FuzzSuccess(f *testing.F) {
	// works as 'seed'. Used for corner-case values, valid minimal examples, or previously-found crashes.
	f.Add(200)

	go func() {
		f.Fuzz(func(t *testing.T, i int) {
			d := double(i)
			assert.Equal(t, i*i, d, "Input is %d", i)
		})
	}()

	select {
	case <-time.After(10 * time.Second):
		// SkipNow remains an exit code of the fuzzing
		f.SkipNow()
	}

	// Chdir is used to specify fuzz test working directory
	// f.Chdir()
}

// Fuzz runs single time es entrypoint then it run 12 parallel processes
// Entrypoint (first run) doesn't run f.Fuzz method. This method is called 12 times in that processes
// File will be opened 13 times.
func FuzzSuccess2(f *testing.F) {
	// instruction: delete the file before running test to see proper results
	var file, _ = os.OpenFile("testdata/output.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	_, _ = file.WriteString(fmt.Sprintf("File opened by %s\n", f.Name()))
	counter := atomic.Int64{}
	defer func() {
		_, _ = file.WriteString(fmt.Sprintf("Processed %d lines\n", counter.Load()))
	}()

	go func() {
		f.Fuzz(func(t *testing.T, i int) {
			counter.Add(1)
		})
	}()

	select {
	case <-time.After(5 * time.Second):
		f.SkipNow()
	}
}

func FuzzFail(f *testing.F) {
	// let all the fuzzes start working
	time.Sleep(300 * time.Millisecond)

	go func() {
		f.Fuzz(func(t *testing.T, i int) {
			d := double(i)
			// wrong expected value
			// only one error will be outputted
			require.Equal(t, i*i*i, d, "Input is %d", i)
		})
	}()

	select {
	case <-time.After(5 * time.Second):
		f.SkipNow()
	}
}

func double(value int) int {
	return value * value
}
