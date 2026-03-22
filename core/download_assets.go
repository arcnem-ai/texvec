package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

func downloadAsset(modelName, path, downloadURL string) error {
	fmt.Printf("Downloading %s for %s...\n", filepath.Base(path), modelName)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected status %s", downloadURL, resp.Status)
	}

	tmpPath := path + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer out.Close()

	bar := progressbar.NewOptions64(
		resp.ContentLength,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionClearOnFinish(),
	)

	if _, err := io.Copy(io.MultiWriter(out, bar), resp.Body); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
