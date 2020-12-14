package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
)

type Stream struct {
	v       string
	command *exec.Cmd
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
}

func NewStream(v string) (*Stream, error) {
	s := &Stream{
		v: v,
	}
	return s, nil
}

func prepareTestDirTree() (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v\n", err)
	}
	return tmpDir, nil
}
