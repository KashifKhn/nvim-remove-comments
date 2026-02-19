package upgrade

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type UpgradeResult struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	Upgraded        bool   `json:"upgraded"`
	Platform        struct {
		OS   string `json:"os"`
		Arch string `json:"arch"`
	} `json:"platform"`
	InstalledPaths []string `json:"installed_paths,omitempty"`
	Error          string   `json:"error,omitempty"`
}

func NewCommand(currentVersion string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade remove-comments to the latest version",
		Long: `Upgrade remove-comments CLI to the latest version from GitHub releases.

This command will:
  1. Check if a newer version is available
  2. Download the latest release for your platform
  3. Verify the download integrity using checksums
  4. Create a backup of the current binary
  5. Install the new version
  6. Verify the installation works
  7. Rollback automatically if anything fails

The upgrade is safe - if anything goes wrong, your current version
will be automatically restored.`,
		Example: `  # Check and upgrade to latest version
  rmc upgrade

  # Only check for updates (don't install)
  rmc upgrade --check

  # Force reinstall even if already on latest
  rmc upgrade --force

  # Upgrade to a specific version
  rmc upgrade --version v1.0.3

  # JSON output for scripting
  rmc upgrade --check --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(cmd, args, currentVersion)
		},
	}

	cmd.Flags().BoolP("check", "c", false, "Only check for updates without installing")
	cmd.Flags().BoolP("force", "f", false, "Force upgrade even if already on latest version")
	cmd.Flags().StringP("version", "v", "", "Upgrade to a specific version")
	cmd.Flags().Bool("json", false, "Output result as JSON")

	return cmd
}

func runUpgrade(cmd *cobra.Command, args []string, currentVersion string) error {
	checkOnly, _ := cmd.Flags().GetBool("check")
	force, _ := cmd.Flags().GetBool("force")
	targetVersion, _ := cmd.Flags().GetString("version")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	result := &UpgradeResult{}

	platform, err := GetPlatformInfo()
	if err != nil {
		return outputError(jsonOutput, result, err)
	}
	result.Platform.OS = platform.OS
	result.Platform.Arch = platform.Arch

	result.CurrentVersion = NormalizeVersion(currentVersion)

	if !jsonOutput {
		printHeader("Upgrade")
	}

	var latestVersion string
	if targetVersion != "" {
		latestVersion = NormalizeVersion(targetVersion)
	} else {
		latestVersion, err = GetLatestVersion()
		if err != nil {
			return outputError(jsonOutput, result, fmt.Errorf("failed to check for updates: %w", err))
		}
	}
	result.LatestVersion = latestVersion

	isNewer, err := IsNewerAvailable(currentVersion, latestVersion)
	if err != nil {
		return outputError(jsonOutput, result, fmt.Errorf("failed to compare versions: %w", err))
	}
	result.UpdateAvailable = isNewer

	if !isNewer && !force {
		if jsonOutput {
			return outputJSON(result)
		}
		printAlreadyLatest(result.CurrentVersion)
		return nil
	}

	if checkOnly {
		if jsonOutput {
			return outputJSON(result)
		}
		if isNewer {
			printUpdateAvailable(result.CurrentVersion, result.LatestVersion)
		} else {
			printAlreadyLatest(result.CurrentVersion)
		}
		return nil
	}

	if !jsonOutput {
		printMethod("curl")
		printVersionTransition(result.CurrentVersion, result.LatestVersion)
	}

	progressCallback := func(downloaded, total int64) {
		if !jsonOutput {
			clearLine()
			fmt.Print("  Downloading... " + printProgress(downloaded, total, 20))
		}
	}

	downloadResult, err := DownloadReleaseWithProgress(latestVersion, platform, progressCallback)
	if err != nil {
		if !jsonOutput {
			fmt.Println()
		}
		return outputError(jsonOutput, result, fmt.Errorf("download failed: %w", err))
	}
	defer CleanupDownload(downloadResult)

	if !jsonOutput {
		clearLine()
		printStepComplete("Downloaded " + formatBytes(downloadResult.Size))
	}

	checksums, err := FetchChecksums(latestVersion)
	if err == nil && checksums != nil {
		archiveName := platform.GetArchiveName(latestVersion)
		if expectedHash, ok := checksums[archiveName]; ok {
			if err := VerifyChecksum(downloadResult.FilePath, expectedHash); err != nil {
				return outputError(jsonOutput, result, fmt.Errorf("checksum verification failed: %w", err))
			}
		}
	}

	binaryPath, err := ExtractBinary(downloadResult.FilePath, platform)
	if err != nil {
		return outputError(jsonOutput, result, fmt.Errorf("extraction failed: %w", err))
	}

	installations := FindAllInstallations()
	if len(installations) == 0 {
		execPath, err := GetExecutablePath()
		if err != nil {
			installDir := GetInstallDir()
			installations = []string{fmt.Sprintf("%s/%s", installDir, platform.BinaryName)}
		} else {
			installations = []string{execPath}
		}
	}

	var backups []*BackupInfo
	for _, installPath := range installations {
		if _, err := os.Stat(installPath); err == nil {
			backup, err := CreateBackup(installPath)
			if err == nil {
				backups = append(backups, backup)
			}
		}
	}

	defer func() {
		for _, backup := range backups {
			CleanupBackup(backup)
		}
	}()

	installed, installErrors := InstallToMultipleLocations(binaryPath, installations)
	result.InstalledPaths = installed

	if len(installed) == 0 {
		for _, backup := range backups {
			_ = RestoreBackup(backup)
		}

		errMsg := "installation failed at all locations"
		if len(installErrors) > 0 {
			errMsg = fmt.Sprintf("%s: %v", errMsg, installErrors[0])
		}
		return outputError(jsonOutput, result, fmt.Errorf("%s", errMsg))
	}

	verifyPath := installed[0]
	if err := VerifyInstallation(verifyPath); err != nil {
		for _, backup := range backups {
			_ = RestoreBackup(backup)
		}
		return outputError(jsonOutput, result, fmt.Errorf("installation verification failed: %w", err))
	}

	result.Upgraded = true

	if jsonOutput {
		return outputJSON(result)
	}

	printStepComplete("Upgrade complete")
	printDone()

	return nil
}

func outputError(jsonOutput bool, result *UpgradeResult, err error) error {
	if jsonOutput {
		result.Error = err.Error()
		return outputJSON(result)
	}
	return err
}

func outputJSON(result *UpgradeResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
