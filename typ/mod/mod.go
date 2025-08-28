package mod

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"strings"

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

	Abs string
	Ver Version

	Host string // github.com
	User string // paraskun
	Repo string // o2
	Path string // std/math, empty if root

	Mods map[string]*Module
	Pkgs map[string]*Package
}

func Load(abs string) (*Module, error) {
	yml, err := os.ReadFile(abs + "/o2.yml")

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("not an o2 module")
		}

		return nil, err
	}

	mod := &Module{
		Abs:  abs,
		Mods: make(map[string]*Module),
		Pkgs: make(map[string]*Package),
	}

	if err := yaml.Unmarshal(yml, mod); err != nil {
		return nil, err
	}

	tok := strings.SplitN(mod.Mod.Name, "/", 4)

	if len(tok) < 3 {
		return nil, fmt.Errorf("malformed module name")
	}

	mod.Host = tok[0]
	mod.User = tok[1]
	mod.Repo = tok[2]

	if len(tok) > 3 {
		mod.Path = tok[3]
	}

	if err := mod.Index(); err != nil {
		return nil, fmt.Errorf("index: %w", err)
	}

	for _, full := range mod.Mod.Uses {
		loc, err := Lookup(full)

		if err != nil {
			return nil, fmt.Errorf("could not find %s: %w", full, err)
		}

		use, err := Load(loc)

		if err != nil {
			return nil, fmt.Errorf("cound not load %s: %w", full, err)
		}

		mod.Mods[full] = use
		maps.Copy(mod.Pkgs, use.Pkgs)
	}

	return mod, nil
}

func (m *Module) Index() error {
	return m.index(os.DirFS(m.Abs), m.Mod.Name, m.Abs)
}

func (m *Module) index(fsys fs.FS, name string, path string) error {
	entries, err := fs.ReadDir(fsys, ".")

	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	pkg := &Package{
		Mod: m,
		Src: make([]*File, 0),
		Nat: make([]*File, 0),
		Sym: make(map[string]*typ.Object),
	}

	for _, ent := range entries {
		if ent.Type().IsDir() {
			sub, err := fs.Sub(fsys, ent.Name())

			if err != nil {
				return fmt.Errorf("walk: %w", err)
			}

			if err := m.index(sub, name+"/"+ent.Name(), path+"/"+ent.Name()); err != nil {
				return err
			}
		}

		if ent.Type().IsRegular() {
			if strings.HasSuffix(ent.Name(), ".o2") {
				pkg.Src = append(pkg.Src, &File{
					Name: ent.Name(),
					Path: path + "/" + ent.Name(),
					Pkg:  pkg,
				})

				continue
			}

			if strings.HasSuffix(ent.Name(), ".c") {
				pkg.Nat = append(pkg.Nat, &File{
					Name: ent.Name(),
					Path: path + "/" + ent.Name(),
					Pkg:  pkg,
				})

				continue
			}
		}
	}

	if len(pkg.Src) != 0 {
		m.Pkgs[name] = pkg
	}

	return nil
}

func Lookup(full string) (string, error) {
	tok := strings.Split(full, "@")

	n := tok[0]
	v := tok[1]

	dir, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("could not access user home: %w", err)
	}

	dir += "/o2/mod/" + n + "/" + v

	return dir, nil
}

func (m *Module) Lookup(name string) *Package {
	return m.Pkgs[name]
}

type Package struct {
	Path string // cmd/ast

	Mod *Module
	Src []*File
	Nat []*File

	// Typing pass

	Env *typ.Env
	Sym map[string]*typ.Object
}

func (p *Package) Index() error {
	return nil
}

type File struct {
	Name string // ast.x
	Path string // filesystem absolute path

	Pkg *Package

	// Typing pass

	Dec any      // *ast.File, import cycle otherwise
	Env *typ.Env // Pkg.Env
}
