package polymorphichelpers

type RolloutViewer interface {
	Describe() (string, error)
}
