package typ

type Package struct {
	Name  string
	Usage map[string]*Package
}

// Native is package written in Go.
type Native struct {
	Package
}
