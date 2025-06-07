package curlyraft

type StateMachine interface {
	Apply(command []byte) (result []byte, err error)
}
