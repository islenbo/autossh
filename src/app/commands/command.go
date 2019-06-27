package commands

type Command interface {
	Process() bool
}
