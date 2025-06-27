package typ

type File struct {
	Name string // ast.x

	Env *Env
	Pkg *Package
	Use []*Package
}

type Package struct {
	Path string // ast
}

type Module struct {
	Name string // github.com/paraskun/x
}
