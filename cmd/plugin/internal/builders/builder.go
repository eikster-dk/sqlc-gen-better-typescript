package builders

type File struct {
	Name    string
	Content []byte
}

type Builder interface {
	Build() ([]File, error)
}
