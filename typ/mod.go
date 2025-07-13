package typ

import (
	"fmt"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d",
		v.Major,
		v.Minor,
		v.Patch,
	)
}

func (v *Version) Parse(s string) error {
	_, err := fmt.Sscanf(s, "v%d.%d.%d", &v.Major, &v.Minor, &v.Patch)

	if err != nil {
		return err
	}

	return nil
}

type Module struct {
	Host string // github.com
	User string // paraskun
	Repo string // x
	Path string // mod/math, empty if root

	Version Version

	Pkg map[string]*Package
	Use map[string]*Module
}

func Import(path string) *Module

func (mod *Module) Lookup(path string) (*Package, error)

type Package struct {
	Name string // ast
	Path string // cmd/ast, empty if root

	Mod *Module

	XSrc []*File
	CSrc []*File

	// Typing pass

	Env *Env
	Imm map[string]*Object
}

func (pkg *Package) Insert(lit string, typ *Type) *Object {
	obj, ok := pkg.Imm[lit]

	if !ok {
		obj = &Object{typ, nil}
		pkg.Imm[lit] = obj
	}

	return obj
}

type File struct {
	Name string // ast.x
	Path string // filesystem absolute path

	Pkg *Package

	// Typing pass

	Dec any  // *ast.File, import cycle otherwise
	Env *Env // Pkg.Env
}
