package imagepull

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// LoadTarballImage loads a flow image from a tarball, either from a URL or a local file path, and extracts it to outputDir.
func LoadTarballImage(source string, outputDir string) error {
	var reader io.ReadCloser
	var err error
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		resp, err := http.Get(source)
		if err != nil {
			return fmt.Errorf("failed to download tarball: %w", err)
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("failed to download tarball: status %d", resp.StatusCode)
		}
		reader = resp.Body
	} else {
		reader, err = os.Open(source)
		if err != nil {
			return fmt.Errorf("failed to open tarball: %w", err)
		}
	}
	defer reader.Close()

	// Handle gzip compression if needed
	if strings.HasSuffix(source, ".gz") {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	tr := tar.NewReader(reader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tarball: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue // skip non-regular files
		}
		if strings.Contains(hdr.Name, "..") {
			slog.Warn("Skipping file as it uses '..' within the path. This is not allowed to prevent path traversal attacks.", "name", hdr.Name)
			continue
		}
		outPath := filepath.Join(outputDir, hdr.Name)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		outFile, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}
		outFile.Close()
	}
	return nil
}
