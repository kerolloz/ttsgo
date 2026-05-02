package logger

import (
	"bytes"
	"regexp"
	"testing"
)

func captureOutput(fn func()) (stdout, stderr string) {
	var outBuf, errBuf bytes.Buffer
	orig, origErr := Stdout, Stderr
	Stdout, Stderr = &outBuf, &errBuf
	defer func() { Stdout, Stderr = orig, origErr }()
	fn()
	return outBuf.String(), errBuf.String()
}

var timestampRe = regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\]`)

func TestInfo(t *testing.T) {
	out, err := captureOutput(func() { Info("hello %s", "world") })
	if err != "" {
		t.Errorf("Info wrote to stderr: %q", err)
	}
	if !timestampRe.MatchString(out) {
		t.Errorf("missing timestamp in %q", out)
	}
	for _, want := range []string{"INFO", "hello world", "[Nego]"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("Info output missing %q: %q", want, out)
		}
	}
}

func TestWarn(t *testing.T) {
	out, _ := captureOutput(func() { Warn("something") })
	if !bytes.Contains([]byte(out), []byte("WARN")) {
		t.Errorf("Warn output missing WARN: %q", out)
	}
}

func TestError(t *testing.T) {
	out, err := captureOutput(func() { Error("bad thing") })
	if out != "" {
		t.Errorf("Error wrote to stdout: %q", out)
	}
	if !bytes.Contains([]byte(err), []byte("ERROR")) {
		t.Errorf("Error output missing ERROR: %q", err)
	}
	if !bytes.Contains([]byte(err), []byte("bad thing")) {
		t.Errorf("Error output missing message: %q", err)
	}
}

func TestSuccess(t *testing.T) {
	out, _ := captureOutput(func() { Success("done") })
	if !bytes.Contains([]byte(out), []byte("READY")) {
		t.Errorf("Success output missing READY: %q", out)
	}
}

func TestStep(t *testing.T) {
	out, _ := captureOutput(func() { Step("watching") })
	if !bytes.Contains([]byte(out), []byte("STEP")) {
		t.Errorf("Step output missing STEP: %q", out)
	}
}
