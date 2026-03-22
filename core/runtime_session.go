package core

import (
	"fmt"

	onnxruntime "github.com/yalue/onnxruntime_go"
)

type RuntimeOptions struct {
	Verbose bool
}

func ensureONNXRuntime() error {
	if onnxruntime.IsInitialized() {
		return nil
	}

	if err := EnsureRuntime(); err != nil {
		return err
	}

	return onnxruntime.InitializeEnvironment()
}

func newSessionOptions(options RuntimeOptions) (*onnxruntime.SessionOptions, error) {
	opts, err := onnxruntime.NewSessionOptions()
	if err != nil {
		return nil, err
	}

	if options.Verbose {
		fmt.Println("Using CPU execution provider")
	}

	return opts, nil
}
