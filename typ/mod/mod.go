package mod

import (
	"errors"
	"fmt"
	"os"

	"github.com/paraskun/o2/typ"

	yaml "gopkg.in/yaml.v3"
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
	Mod struct {
		Name string
		Uses []string
	}

	Host string // github.com
	User string // paraskun
	Repo string // o2
	Path string // std/math, empty if root

	Version Version

	Pkg map[string]*Package
}

func (mod *Module) Load(path string) error {
	yml, err := os.ReadFile(path + "/x.yml")

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not inside o2 module")
		}

		return err
	}

	if err := yaml.Unmarshal(yml, mod); err != nil {
		return err
	}

	return nil
}

func (mod *Module) Fetch() (string, error) {
	return "", nil
}

func (mod *Module) Lookup(path string) *Package {
	return nil
}

type Package struct {
	Name string // ast
	Path string // cmd/ast, empty if root

	Mod *Module

	XSrc []*File
	CSrc []*File

	// Typing pass

	Env *typ.Env
	Imm map[string]*typ.Object
}

func (pkg *Package) Insert(l string, t *typ.Type) *typ.Object {
	obj, ok := pkg.Imm[l]

	if !ok {
		obj = &typ.Object{
			Typ: t,
			Seg: typ.PDat,
		}
		pkg.Imm[l] = obj
	}

	return obj
}

type File struct {
	Name string // ast.x
	Path string // filesystem absolute path

	Pkg *Package

	// Typing pass

	Dec any      // *ast.File, import cycle otherwise
	Env *typ.Env // Pkg.Env
}
