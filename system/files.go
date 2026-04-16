package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const QuadletDir = ".config/containers/systemd"

func GetQuadletDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, QuadletDir), nil
}

type QuadletFile struct {
	Name    string
	Path    string
	Type    string
	Content string
}

func ListQuadletFiles() ([]QuadletFile, error) {
	dir, err := GetQuadletDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []QuadletFile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading quadlet directory: %w", err)
	}

	var files []QuadletFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		if !isValidQuadletType(ext) {
			continue
		}

		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		files = append(files, QuadletFile{
			Name:    strings.TrimSuffix(name, ext),
			Path:    path,
			Type:    strings.TrimPrefix(ext, "."),
			Content: string(content),
		})
	}

	return files, nil
}

func ReadQuadletFile(name string) (QuadletFile, error) {
	dir, err := GetQuadletDir()
	if err != nil {
		return QuadletFile{}, err
	}

	for _, ext := range []string{".container", ".pod", ".kube", ".volume", ".network"} {
		path := filepath.Join(dir, name+ext)
		content, err := os.ReadFile(path)
		if err == nil {
			return QuadletFile{
				Name:    name,
				Path:    path,
				Type:    strings.TrimPrefix(ext, "."),
				Content: string(content),
			}, nil
		}
	}

	return QuadletFile{}, fmt.Errorf("quadlet file not found: %s", name)
}

func WriteQuadletFile(name string, ext string, content string) error {
	dir, err := GetQuadletDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating quadlet directory: %w", err)
	}

	path := filepath.Join(dir, name+ext)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing quadlet file: %w", err)
	}

	return nil
}

func DeleteQuadletFile(name string) error {
	dir, err := GetQuadletDir()
	if err != nil {
		return err
	}

	deleted := false
	for _, ext := range []string{".container", ".pod", ".kube", ".volume", ".network"} {
		path := filepath.Join(dir, name+ext)
		if err := os.Remove(path); err == nil {
			deleted = true
		}
	}

	if !deleted {
		return fmt.Errorf("no quadlet files found for: %s", name)
	}

	return nil
}

func isValidQuadletType(ext string) bool {
	switch ext {
	case ".container", ".pod", ".kube", ".volume", ".network":
		return true
	default:
		return false
	}
}
