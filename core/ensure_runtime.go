package core

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arcnem-ai/texvec/config"
	"github.com/schollz/progressbar/v3"
	onnxruntime "github.com/yalue/onnxruntime_go"
)

const onnxRuntimeVersion = "1.24.3"

func runtimeDownloadURL() (string, error) {
	base := "https://github.com/microsoft/onnxruntime/releases/download/v" + onnxRuntimeVersion

	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			return base + "/onnxruntime-osx-arm64-" + onnxRuntimeVersion + ".tgz", nil
		case "amd64":
			return base + "/onnxruntime-osx-x86_64-" + onnxRuntimeVersion + ".tgz", nil
		}
	case "linux":
		switch runtime.GOARCH {
		case "arm64":
			return base + "/onnxruntime-linux-aarch64-" + onnxRuntimeVersion + ".tgz", nil
		case "amd64":
			return base + "/onnxruntime-linux-x64-" + onnxRuntimeVersion + ".tgz", nil
		}
	}

	return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func runtimeLibName() string {
	if runtime.GOOS == "darwin" {
		return "libonnxruntime.dylib"
	}

	return "libonnxruntime.so"
}

func RuntimeLibPath() (string, error) {
	libDir, err := config.RuntimeLibDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(libDir, runtimeLibName()), nil
}

func EnsureRuntime() error {
	libPath, err := RuntimeLibPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(libPath); err == nil {
		onnxruntime.SetSharedLibraryPath(libPath)
		return nil
	}

	url, err := runtimeDownloadURL()
	if err != nil {
		return err
	}

	fmt.Printf("Downloading ONNX Runtime %s...\n", onnxRuntimeVersion)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download ONNX Runtime: %s", resp.Status)
	}

	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionClearOnFinish(),
	)

	pr, pw := io.Pipe()
	go func() {
		_, err := io.Copy(io.MultiWriter(pw, bar), resp.Body)
		pw.CloseWithError(err)
	}()

	gz, err := gzip.NewReader(pr)
	if err != nil {
		return fmt.Errorf("decompress: %w", err)
	}
	defer gz.Close()

	versionedName := "libonnxruntime." + onnxRuntimeVersion + ".dylib"
	if runtime.GOOS == "linux" {
		versionedName = "libonnxruntime.so." + onnxRuntimeVersion
	}

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("ONNX Runtime library not found in archive")
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}

		if filepath.Base(hdr.Name) != versionedName || hdr.Typeflag != tar.TypeReg {
			continue
		}

		tmpPath := libPath + ".tmp"
		out, err := os.Create(tmpPath)
		if err != nil {
			return err
		}

		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			os.Remove(tmpPath)
			return err
		}
		out.Close()

		if err := os.Rename(tmpPath, libPath); err != nil {
			return err
		}

		fmt.Println("ONNX Runtime installed.")
		onnxruntime.SetSharedLibraryPath(libPath)
		return nil
	}
}
