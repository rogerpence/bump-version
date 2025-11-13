 I use this Go program when creating private PNPM packages. These packages generally need unique tags to identify unique builds.

 The workflow is simple:

 - Make changes to a PNPM package project. 
 - Run `bump-version [commit message]`

`bump-version`:

- increments the minor value of the version number
- stages all files 
- commits changes
- creates a new tag with the new version number
- commits the tag
- pushes the commit and the new tag



# Run directly

```
go run bump-version.go
```

# Or build and run

```
go build bump-version.go
.\bump-version.exe
```

Go lets you build little utilities like this very quickly--which Python and PowerShell also do. Go's ace up its sleeve is that it makes it dead simple to create and deploy a Go executable. Put the executable in directory in your path and `bump-version` is always available.
