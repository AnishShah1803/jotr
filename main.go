package main

import (
	"os"

	"github.com/AnishShah1803/jotr/cmd"
	"github.com/AnishShah1803/jotr/internal/utils"
)

func main() {
	if err := cmd.Execute(); err != nil {
		utils.PrintError("%v", err)
		os.Exit(1)
	}
}
