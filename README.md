This Go program makes it easy to increment a package version number and promote that number to GitHub. 

 The workflow is simple:

 - Make changes to a PNPM package project. 
 - Run `bump-version [commit message]`

`bump-version`:

- increments the version number
- stages all files 
- commits changes
- creates a new tag with the new version number
- commits the tag
- pushes the commit and the new tag

Because this utility is aimed primarily at managing project version numbers for GitHub-hosted Node packages, it also: 

- creates a PNPM command line like this `pnpm add https://github.com/rogerpence/rp-utils#v1.0.16` and puts it on the clipboard. This makes it easy to update consuming projects. 
- This command line is also written to the console.

The GitHub account name is hardcoded in the program's `githubAccount` variable. 

# Example
```
bump-version [--dryrun] <commit comments>
```

The commit comments are required. `--dryrun` shows the results without changing anything.

# Run directly

```
go run bump-version.go
```

# Or build and run

```
go build bump-version.go
.\bump-version.exe

```

## Copy bump-version.exe to a utilities folder

The `deploy.ps1` PowerShell script is for copying the `bump-version.exe` to a utilities folder that is in Windows' path. 


Go lets you build little utilities like this very quickly--which Python, Node, and PowerShell also do. Go's ace up its sleeve is that it makes it dead simple to create and deploy a Go executable. Put the executable in directory in your path and `bump-version` is always available. 

(PS: I wrote this utility just before I started using Bun (which can compile Typescript to an executable). Oh, well, I got to play with Go doing this.


