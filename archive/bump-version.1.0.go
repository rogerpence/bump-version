package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const filePermissions = 0644 // rw-r--r--

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

	// Parse JSON to get version and name (for validation and clipboard)
	var pkg map[string]interface{}
	err = json.Unmarshal(data, &pkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package.json: %v\n", err)
		os.Exit(1)
	}

	// Get version
	version, ok := pkg["version"].(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: version field not found or not a string\n")
		os.Exit(1)
	}

	// Parse version
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "Invalid version format: %s (expected x.y.z)\n", version)
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
	oldVersion := version

	fmt.Printf("Bumping version: %s -> %s\n", oldVersion, newVersion)

	// Replace version in the original JSON string to preserve key order
	versionPattern := regexp.MustCompile(`("version"\s*:\s*)"` + regexp.QuoteMeta(oldVersion) + `"`)
	updatedData := versionPattern.ReplaceAll(data, []byte(`${1}"`+newVersion+`"`))

	err = os.WriteFile(packageFile, updatedData, filePermissions)
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
	name, _ := pkg["name"].(string)
	packageName := strings.TrimPrefix(name, "@")
	clipboardText := fmt.Sprintf(clipboardMask, packageName, newVersion)

	// Copy to clipboard
	cmd = exec.Command("pwsh", "-command", fmt.Sprintf("Set-Clipboard -Value '%s'", clipboardText))
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not copy to clipboard: %v\n", err)
	} else {
		fmt.Printf("\n📋 Copied to clipboard: %s\n", clipboardText)
	}
}
