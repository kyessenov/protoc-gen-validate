package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/protoc-gen-validate/tests/harness"
	"golang.org/x/net/context"
)

var Harnesses = []Harness{
	InitHarness("tests/harness/go/go-harness"),
}

type Harness func(context.Context, io.Reader) (*harness.TestResult, error)

func InitHarness(cmd string, args ...string) Harness {
	return func(ctx context.Context, r io.Reader) (*harness.TestResult, error) {
		out, errs := getBuf(), getBuf()
		defer relBuf(out)
		defer relBuf(errs)

		cmd := exec.CommandContext(ctx, cmd, args...)
		cmd.Stdin = r
		cmd.Stdout = out
		cmd.Stderr = errs

		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("[%s] failed execution (%v) - captured stderr:\n%s", cmdStr(cmd), err, errs.String())
		}

		res := new(harness.TestResult)
		if err := proto.Unmarshal(out.Bytes(), res); err != nil {
			return nil, fmt.Errorf("[%s] failed to unmarshal result: %v", cmdStr(cmd), err)
		}

		return res, nil
	}
}

var bufPool = &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}

func getBuf() *bytes.Buffer {
	return bufPool.Get().(*bytes.Buffer)
}

func relBuf(b *bytes.Buffer) {
	b.Reset()
	bufPool.Put(b)
}

func cmdStr(cmd *exec.Cmd) string {
	return fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
}
