package main

import "github.com/KashifKhn/remove-comments/cli/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
