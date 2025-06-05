package api

type StateMachine interface {
	Apply(command []byte) (result []byte, err error)
}
