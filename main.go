package main

import "github.com/ainsuotain/sshtie/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
