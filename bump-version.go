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

const githubAccount = "rogerpence"

type VersionConfig struct {
	BumpType   string
	DryRun     bool
	CommitMsg  string
	OldVersion string
	NewVersion string
}

type PackageInfo struct {
	Version          string
	Name             string
	HasPackageScript bool
}

func parseCommandLineArgs() (bumpType string, dryRun bool, commitMsg string) {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--major|--minor] [--dryrun] <commit-message>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  --major:   Bump major version (x.0.0)\n")
		fmt.Fprintf(os.Stderr, "  --minor:   Bump minor version (x.y.0)\n")
		fmt.Fprintf(os.Stderr, "  --dryrun:  Show what would happen without making changes\n")
		fmt.Fprintf(os.Stderr, "  (default)  Bump patch version (x.y.z)\n")
		os.Exit(1)
	}

	bumpType = "patch"
	dryRun = false
	commitMsg = ""

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--major" || arg == "-major" {
			bumpType = "major"
		} else if arg == "--minor" || arg == "-minor" {
			bumpType = "minor"
		} else if arg == "--dryrun" || arg == "-dryrun" {
			dryRun = true
		} else {
			commitMsg = arg
			break
		}
	}

	if commitMsg == "" {
		fmt.Fprintf(os.Stderr, "Error: commit message is required\n")
		os.Exit(1)
	}

	return
}

func readPackageJSON() (data []byte, pkg map[string]interface{}) {
	packageFile := "package.json"

	data, err := os.ReadFile(packageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading package.json: %v\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(data, &pkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package.json: %v\n", err)
		os.Exit(1)
	}

	return
}

func getPackageInfo(pkg map[string]interface{}) PackageInfo {
	version, ok := pkg["version"].(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: version field not found or not a string\n")
		os.Exit(1)
	}

	name, _ := pkg["name"].(string)

	hasPackageScript := false
	if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
		_, hasPackage := scripts["package"]
		_, hasPrepackage := scripts["prepack"]
		hasPackageScript = hasPackage || hasPrepackage
	}

	return PackageInfo{
		Version:          version,
		Name:             name,
		HasPackageScript: hasPackageScript,
	}
}

func calculateNewVersion(oldVersion, bumpType string) string {
	parts := strings.Split(oldVersion, ".")
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "Invalid version format: %s (expected x.y.z)\n", oldVersion)
		os.Exit(1)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing major version: %v\n", err)
		os.Exit(1)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing minor version: %v\n", err)
		os.Exit(1)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing patch version: %v\n", err)
		os.Exit(1)
	}

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func updatePackageJSON(data []byte, oldVersion, newVersion string, dryRun bool) {
	versionPattern := regexp.MustCompile(`("version"\s*:\s*)"` + regexp.QuoteMeta(oldVersion) + `"`)
	updatedData := versionPattern.ReplaceAll(data, []byte(`${1}"`+newVersion+`"`))

	if !dryRun {
		err := os.WriteFile("package.json", updatedData, filePermissions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing package.json: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Updated package.json to version %s\n", newVersion)
	} else {
		fmt.Printf("Would update package.json to version %s\n", newVersion)
	}
}

func runPackageScript(dryRun bool) {
	if !dryRun {
		fmt.Println("\n📦 Running 'pnpm run package'...")
		cmd := exec.Command("pnpm", "run", "package")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running pnpm run package: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Package built successfully")
	} else {
		fmt.Println("Would run: pnpm run package")
	}
}

func gitCommitAndPush(commitMsg, tagName string, newVersion string, dryRun bool) {
	if !dryRun {
		// Git stage all files
		cmd := exec.Command("git", "add", "-A")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running git add: %v\n", err)
			os.Exit(1)
		}

		// Git commit
		cmd = exec.Command("git", "commit", "-m", commitMsg)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running git commit: %v\n%s\n", err, output)
			os.Exit(1)
		}
		fmt.Println("✓ Committed changes")

		// Git tag
		cmd = exec.Command("git", "tag", tagName)
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating git tag: %v\n%s\n", err, output)
			os.Exit(1)
		}

		fmt.Printf("✓ Created git tag %s\n", tagName)

		// Git push
		cmd = exec.Command("git", "push")
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running git push: %v\n%s\n", err, output)
			os.Exit(1)
		}

		fmt.Println("✓ Pushed commits to remote")

		// Git push tags
		cmd = exec.Command("git", "push", "--tags")
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error pushing tags: %v\n%s\n", err, output)
			os.Exit(1)
		}

		fmt.Println("✓ Pushed tags to remote")
		fmt.Printf("\n✅ Successfully bumped to version %s and pushed to remote!\n", newVersion)
	} else {
		fmt.Println("\nWould execute:")
		fmt.Println("  git add -A")
		fmt.Printf("  git commit -m \"%s\"\n", commitMsg)
		fmt.Printf("  git tag %s\n", tagName)
		fmt.Println("  git push")
		fmt.Println("  git push --tags")
		fmt.Printf("\n✅ Dry run complete - version would be %s\n", newVersion)
	}
}

func copyInstallCommandToClipboard(packageName, newVersion string, dryRun bool) {
	clipboardMask := fmt.Sprintf("pnpm add https://github.com/%s/%%s#v%%s", githubAccount)
	packageName = strings.TrimPrefix(packageName, "@")
	clipboardText := fmt.Sprintf(clipboardMask, packageName, newVersion)

	if !dryRun {
		cmd := exec.Command("pwsh", "-command", fmt.Sprintf("Set-Clipboard -Value '%s'", clipboardText))
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not copy to clipboard: %v\n", err)
		} else {
			fmt.Printf("\n📋 Copied to clipboard: %s\n", clipboardText)
		}
	} else {
		fmt.Printf("\nWould copy to clipboard: %s\n", clipboardText)
	}
}

func main() {
	// Parse command line arguments
	bumpType, dryRun, commitMsg := parseCommandLineArgs()

	// Read and parse package.json
	data, pkg := readPackageJSON()
	pkgInfo := getPackageInfo(pkg)

	// Calculate new version
	newVersion := calculateNewVersion(pkgInfo.Version, bumpType)
	oldVersion := pkgInfo.Version

	fmt.Printf("Bumping version: %s -> %s\n", oldVersion, newVersion)

	if dryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No changes will be made\n")
	}

	// Update package.json with new version
	updatePackageJSON(data, oldVersion, newVersion, dryRun)

	// Run package script if it exists
	if pkgInfo.HasPackageScript {
		runPackageScript(dryRun)
	}

	// Commit and push changes with tag
	tagName := fmt.Sprintf("v%s", newVersion)
	gitCommitAndPush(commitMsg, tagName, newVersion, dryRun)

	// Copy install command to clipboard
	copyInstallCommandToClipboard(pkgInfo.Name, newVersion, dryRun)
}
