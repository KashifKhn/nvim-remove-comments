package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	BinaryName = "remove-comments"
	RepoOwner  = "KashifKhn"
	RepoName   = "remove-comments"
)

type PlatformInfo struct {
	OS         string
	Arch       string
	BinaryName string
	ArchiveExt string
}

func GetPlatformInfo() (*PlatformInfo, error) {
	osName := GetOS()
	arch := GetArch()

	if osName == "unknown" {
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	if arch == "unknown" {
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	info := &PlatformInfo{
		OS:         osName,
		Arch:       arch,
		BinaryName: BinaryName,
		ArchiveExt: ".tar.gz",
	}

	if osName == "windows" {
		info.BinaryName = BinaryName + ".exe"
		info.ArchiveExt = ".zip"
	}

	return info, nil
}

func GetOS() string {
	switch runtime.GOOS {
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	case "windows":
		return "windows"
	default:
		return "unknown"
	}
}

func GetArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return "unknown"
	}
}

func GetExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	return execPath, nil
}

func GetInstallDir() string {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, "Microsoft", "WindowsApps")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Microsoft", "WindowsApps")
	}

	if isWritable("/usr/local/bin") {
		return "/usr/local/bin"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/usr/local/bin"
	}

	return filepath.Join(homeDir, ".local", "bin")
}

func isWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	testFile := filepath.Join(path, ".rmc_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(testFile)
	return true
}

func (p *PlatformInfo) GetArchiveName(version string) string {
	return fmt.Sprintf("%s-%s-%s%s", BinaryName, p.OS, p.Arch, p.ArchiveExt)
}

func (p *PlatformInfo) GetBinaryNameInArchive() string {
	if p.OS == "windows" {
		return fmt.Sprintf("%s.exe", BinaryName)
	}
	return BinaryName
}

func (p *PlatformInfo) GetDownloadURL(version string) string {
	archiveName := p.GetArchiveName(version)
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		RepoOwner, RepoName, version, archiveName)
}

func (p *PlatformInfo) GetChecksumsURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/checksums.txt",
		RepoOwner, RepoName, version)
}

func (p *PlatformInfo) String() string {
	return fmt.Sprintf("%s/%s", p.OS, p.Arch)
}

func FindAllInstallations() []string {
	var installations []string
	binaryName := BinaryName
	if runtime.GOOS == "windows" {
		binaryName = BinaryName + ".exe"
	}

	searchPaths := getSearchPaths()

	for _, dir := range searchPaths {
		fullPath := filepath.Join(dir, binaryName)
		if _, err := os.Stat(fullPath); err == nil {
			realPath, err := filepath.EvalSymlinks(fullPath)
			if err == nil {
				fullPath = realPath
			}
			if !containsPath(installations, fullPath) {
				installations = append(installations, fullPath)
			}
		}
	}

	return installations
}

func getSearchPaths() []string {
	var paths []string

	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			paths = append(paths, filepath.Join(localAppData, "Microsoft", "WindowsApps"))
		}
	} else {
		paths = append(paths, "/usr/local/bin")

		homeDir, err := os.UserHomeDir()
		if err == nil {
			paths = append(paths, filepath.Join(homeDir, ".local", "bin"))
		}
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			gopath = filepath.Join(homeDir, "go")
		}
	}
	if gopath != "" {
		paths = append(paths, filepath.Join(gopath, "bin"))
	}

	pathEnv := os.Getenv("PATH")
	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}
	for _, p := range strings.Split(pathEnv, pathSeparator) {
		if p != "" && !containsPath(paths, p) {
			paths = append(paths, p)
		}
	}

	return paths
}

func containsPath(paths []string, path string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}
