package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type DownloadResult struct {
	FilePath   string
	Size       int64
	Checksum   string
	BinaryPath string
}

type ProgressCallback func(downloaded, total int64)

func DownloadRelease(version string, platform *PlatformInfo) (*DownloadResult, error) {
	return DownloadReleaseWithProgress(version, platform, nil)
}

func DownloadReleaseWithProgress(version string, platform *PlatformInfo, progress ProgressCallback) (*DownloadResult, error) {
	tmpDir, err := os.MkdirTemp("", "rmc-upgrade-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	archiveName := platform.GetArchiveName(version)
	archivePath := filepath.Join(tmpDir, archiveName)
	downloadURL := platform.GetDownloadURL(version)

	if err := downloadFileWithProgress(downloadURL, archivePath, progress); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to download release: %w", err)
	}

	fileInfo, err := os.Stat(archivePath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to stat downloaded file: %w", err)
	}

	checksum, err := calculateChecksum(archivePath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return &DownloadResult{
		FilePath: archivePath,
		Size:     fileInfo.Size(),
		Checksum: checksum,
	}, nil
}

func downloadFileWithProgress(url, destPath string, progress ProgressCallback) error {
	client := &http.Client{Timeout: RequestTimeout * 10}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("release not found at %s", url)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if progress != nil && resp.ContentLength > 0 {
		var downloaded int64
		buf := make([]byte, 32*1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				_, writeErr := out.Write(buf[:n])
				if writeErr != nil {
					return fmt.Errorf("failed to write file: %w", writeErr)
				}
				downloaded += int64(n)
				progress(downloaded, resp.ContentLength)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}
		}
	} else {
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return nil
}

func calculateChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func FetchChecksums(version string) (map[string]string, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/checksums.txt",
		RepoOwner, RepoName, version)

	client := &http.Client{Timeout: RequestTimeout}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch checksums: status %d", resp.StatusCode)
	}

	checksums := make(map[string]string)
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			filename := parts[len(parts)-1]
			checksums[filename] = hash
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse checksums: %w", err)
	}

	return checksums, nil
}

func VerifyChecksum(filePath, expectedHash string) error {
	actualHash, err := calculateChecksum(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if !strings.EqualFold(actualHash, expectedHash) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func ExtractBinary(archivePath string, platform *PlatformInfo) (string, error) {
	destDir := filepath.Dir(archivePath)
	binaryName := platform.GetBinaryNameInArchive()

	var extractedPath string
	var err error

	if platform.ArchiveExt == ".zip" {
		extractedPath, err = extractFromZip(archivePath, destDir, binaryName)
	} else {
		extractedPath, err = extractFromTarGz(archivePath, destDir, binaryName)
	}

	if err != nil {
		return "", err
	}

	return extractedPath, nil
}

func extractFromTarGz(archivePath, destDir, targetName string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		filename := filepath.Base(header.Name)
		if filename == targetName {
			destPath := filepath.Join(destDir, filename)

			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", fmt.Errorf("failed to extract file: %w", err)
			}
			outFile.Close()

			return destPath, nil
		}
	}

	return "", fmt.Errorf("binary %s not found in archive", targetName)
}

func extractFromZip(archivePath, destDir, targetName string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		filename := filepath.Base(f.Name)
		if filename != targetName {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open file in zip: %w", err)
		}

		destPath := filepath.Join(destDir, filename)
		outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			rc.Close()
			return "", fmt.Errorf("failed to create output file: %w", err)
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return "", fmt.Errorf("failed to extract file: %w", err)
		}

		outFile.Close()
		rc.Close()
		return destPath, nil
	}

	return "", fmt.Errorf("binary %s not found in archive", targetName)
}

func CleanupDownload(result *DownloadResult) {
	if result != nil && result.FilePath != "" {
		os.RemoveAll(filepath.Dir(result.FilePath))
	}
}
