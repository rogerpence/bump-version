package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Copy to clipboard
	clipboardMask := "pnpm add github:%s#v%s"
	newVersion := "1.0.16"
	packageName := "@rogerpence/rp-utils"
	clipboardText := fmt.Sprintf(clipboardMask, packageName, newVersion)

	cmd := exec.Command("pwsh", "-command", fmt.Sprintf("Set-Clipboard -Value '%s'", clipboardText))
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not copy to clipboard: %v\n", err)
	} else {
		fmt.Printf("\n📋 Copied to clipboard: %s\n", clipboardText)
	}
}
