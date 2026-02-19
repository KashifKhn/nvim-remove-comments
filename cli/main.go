package main

import "github.com/KashifKhn/nvim-remove-comments/cli/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
