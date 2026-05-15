package designer

// Project holds the codegen output settings the designer's Settings
// dialog edits. Fields are mutated in place; pass a *Project to Open
// so edits survive across the popup's lifetime.
type Project struct {
	Name     string // shown in the popup header band
	OutPath  string // file written by Save / Generate
	Package  string // emitted package declaration
	FuncName string // emitted func wrapper name
	Theme    string // theme label shown in the header (display only)
}

// DefaultProject returns a Project pre-populated with the values the
// original designer-poc used. Open(..., nil) is equivalent to
// Open(..., DefaultProject()).
func DefaultProject() *Project {
	return &Project{
		Name:     "untitled.go",
		OutPath:  "/tmp/designer-out.go",
		Package:  "main",
		FuncName: "BuildUI",
		Theme:    "TokyoNight",
	}
}
