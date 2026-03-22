package config

import (
	"os"
	"path/filepath"
)

const homeEnvVar = "TEXVEC_HOME"

func BaseDir() (string, error) {
	if override := os.Getenv(homeEnvVar); override != "" {
		if err := os.MkdirAll(override, 0o755); err != nil {
			return "", err
		}
		return override, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, ".texvec")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return dir, nil
}

func DBPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}

	return "file:" + filepath.Join(base, "texvec.db"), nil
}

func ModelsDir() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "models")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return dir, nil
}

func GetModelDir(modelName string) (string, error) {
	modelsDir, err := ModelsDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(modelsDir, modelName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return dir, nil
}

func RuntimeLibDir() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "lib")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return dir, nil
}

func ConfigPath() (string, error) {
	base, err := BaseDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "config.json"), nil
}
