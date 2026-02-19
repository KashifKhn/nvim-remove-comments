package cmd

import (
	"github.com/KashifKhn/remove-comments/cli/internal/upgrade"
)

func registerUpgradeCmd(version string) {
	rootCmd.AddCommand(upgrade.NewCommand(version))
}
