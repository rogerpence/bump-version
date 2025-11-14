package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type PackageJSON struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Type         string                 `json:"type,omitempty"`
	Main         string                 `json:"main,omitempty"`
	Types        string                 `json:"types,omitempty"`
	Exports      map[string]interface{} `json:"exports,omitempty"`
	Files        []string               `json:"files,omitempty"`
	Dependencies map[string]string      `json:"dependencies,omitempty"`
}

func main() {

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <commit-message>\n", os.Args[0])
		os.Exit(1)
	}
	commitMsg := os.Args[1]

	packageFile := "package.json"

	// Read package.json
	data, err := os.ReadFile(packageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading package.json: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var pkg PackageJSON
	err = json.Unmarshal(data, &pkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package.json: %v\n", err)
		os.Exit(1)
	}

	// Parse version
	parts := strings.Split(pkg.Version, ".")
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "Invalid version format: %s (expected x.y.z)\n", pkg.Version)
		os.Exit(1)
	}

	// Increment patch version
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing patch version: %v\n", err)
		os.Exit(1)
	}
	patch++

	// Create new version
	newVersion := fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch)
	oldVersion := pkg.Version
	pkg.Version = newVersion

	fmt.Printf("Bumping version: %s -> %s\n", oldVersion, newVersion)

	// Write updated package.json with pretty formatting
	updatedData, err := json.MarshalIndent(pkg, "", "    ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling package.json: %v\n", err)
		os.Exit(1)
	}

	// Add newline at end of file
	updatedData = append(updatedData, '\n')

	err = os.WriteFile(packageFile, updatedData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing package.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Updated package.json to version %s\n", newVersion)

	// Git stage all files
	cmd := exec.Command("git", "add", "-A")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running git add: %v\n", err)
		os.Exit(1)
	}

	// Git commit
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running git commit: %v\n", err)
		os.Exit(1)
	}

	// Git tag
	tagName := fmt.Sprintf("v%s", newVersion)
	cmd = exec.Command("git", "tag", tagName)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating git tag: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Created git tag %s\n", tagName)

	// Git push
	cmd = exec.Command("git", "push")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running git push: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Pushed commits to remote")

	// Git push tags
	cmd = exec.Command("git", "push", "--tags")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error pushing tags: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Pushed tags to remote")
	fmt.Printf("\n✅ Successfully bumped to version %s and pushed to remote!\n", newVersion)

	// Create clipboard string
	clipboardMask := "pnpm add github:%s#v%s"
	packageName := strings.TrimPrefix(pkg.Name, "@")
	clipboardText := fmt.Sprintf(clipboardMask, packageName, newVersion)

	// Copy to clipboard
	cmd = exec.Command("pwsh", "-command", fmt.Sprintf("Set-Clipboard -Value '%s'", clipboardText))
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not copy to clipboard: %v\n", err)
	} else {
		fmt.Printf("\n📋 Copied to clipboard: %s\n", clipboardText)
	}
}
