package typ

import "fmt"

type File struct {
	Name string // ast.x
	Path string

	Env *Env
	Pkg *Package
}

type Package struct {
	Name string // ast
	Path string // cmd/ast

	Env *Env
	Mod *Module
	Src []*File
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d",
		v.Major,
		v.Minor,
		v.Patch,
	)
}

type Module struct {
	Host []string // github.com/paraskun
	Name string   // x

	Ver Version

	Pkg map[string]*Package
	Use map[string]*Module
}
