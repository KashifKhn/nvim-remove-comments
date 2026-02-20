package upgrade

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type BackupInfo struct {
	OriginalPath string
	BackupPath   string
	CreatedAt    time.Time
}

func CreateBackup(binaryPath string) (*BackupInfo, error) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("binary not found at %s", binaryPath)
	}

	backupDir, err := os.MkdirTemp("", "rmc-backup-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupPath := filepath.Join(backupDir, filepath.Base(binaryPath)+".backup")

	if err := copyFile(binaryPath, backupPath); err != nil {
		_ = os.RemoveAll(backupDir)
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	return &BackupInfo{
		OriginalPath: binaryPath,
		BackupPath:   backupPath,
		CreatedAt:    time.Now(),
	}, nil
}

func RestoreBackup(backup *BackupInfo) error {
	if backup == nil {
		return fmt.Errorf("no backup information provided")
	}

	if _, err := os.Stat(backup.BackupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found at %s", backup.BackupPath)
	}

	if err := copyFile(backup.BackupPath, backup.OriginalPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(backup.OriginalPath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}

	return nil
}

func CleanupBackup(backup *BackupInfo) {
	if backup != nil && backup.BackupPath != "" {
		_ = os.RemoveAll(filepath.Dir(backup.BackupPath))
	}
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		_ = destFile.Close()
		return err
	}

	if err := destFile.Sync(); err != nil {
		_ = destFile.Close()
		return err
	}

	return destFile.Close()
}

func InstallBinary(sourcePath, destPath string) error {
	if runtime.GOOS == "windows" {
		return installBinaryWindows(sourcePath, destPath)
	}
	return installBinaryUnix(sourcePath, destPath)
}

func installBinaryUnix(sourcePath, destPath string) error {
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	oldPath := destPath + ".old"

	if _, err := os.Stat(destPath); err == nil {
		_ = os.Remove(oldPath)
		if err := os.Rename(destPath, oldPath); err != nil {
			return fmt.Errorf("failed to rename old binary: %w", err)
		}
	}

	if err := copyFile(sourcePath, destPath); err != nil {
		if _, statErr := os.Stat(oldPath); statErr == nil {
			_ = os.Rename(oldPath, destPath)
		}
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	_ = os.Remove(oldPath)

	return nil
}

func installBinaryWindows(sourcePath, destPath string) error {
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	oldPath := destPath + ".old"

	if _, err := os.Stat(destPath); err == nil {
		_ = os.Remove(oldPath)
		if err := os.Rename(destPath, oldPath); err != nil {
			return fmt.Errorf("failed to rename old binary: %w", err)
		}
	}

	if err := copyFile(sourcePath, destPath); err != nil {
		if _, statErr := os.Stat(oldPath); statErr == nil {
			_ = os.Rename(oldPath, destPath)
		}
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	_ = os.Remove(oldPath)

	return nil
}

func VerifyInstallation(binaryPath string) error {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found after installation")
	}

	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary verification failed: %w (output: %s)", err, string(output))
	}

	return nil
}

func InstallToMultipleLocations(sourcePath string, locations []string) ([]string, []error) {
	var installed []string
	var errors []error

	for _, destPath := range locations {
		if err := InstallBinary(sourcePath, destPath); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", destPath, err))
		} else {
			installed = append(installed, destPath)
		}
	}

	return installed, errors
}
