package system

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type QuadletResult struct {
	Name    string
	Ext     string
	Content string
}

func ConvertCompose(composeContent string) ([]QuadletResult, error) {
	outputDir, err := os.MkdirTemp("", "quadlet-output-*")
	if err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}
	defer os.RemoveAll(outputDir)

	composeFile, err := os.CreateTemp("", "compose-*.yml")
	if err != nil {
		return nil, fmt.Errorf("creating temp compose file: %w", err)
	}
	defer os.Remove(composeFile.Name())

	if _, err := composeFile.WriteString(composeContent); err != nil {
		return nil, fmt.Errorf("writing compose file: %w", err)
	}
	composeFile.Close()

	cmd := exec.Command("podlet", "--file", outputDir, "compose", composeFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("running podlet: %w\n%s", err, output)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("reading output directory: %w", err)
	}

	var results []QuadletResult
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		if !isValidQuadletType(ext) {
			continue
		}

		path := filepath.Join(outputDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		results = append(results, QuadletResult{
			Name:    strings.TrimSuffix(name, ext),
			Ext:     ext,
			Content: string(content),
		})
	}

	return results, nil
}

func EncodeQuadletData(files []QuadletResult) (string, error) {
	data, err := json.Marshal(files)
	if err != nil {
		return "", fmt.Errorf("encoding quadlet data: %w", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func DecodeQuadletData(encoded string) ([]QuadletResult, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decoding quadlet data: %w", err)
	}

	var files []QuadletResult
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, fmt.Errorf("unmarshaling quadlet data: %w", err)
	}

	return files, nil
}
