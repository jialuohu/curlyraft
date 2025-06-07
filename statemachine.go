package curlyraft

type StateMachine interface {
	Apply(command []byte) (result []byte, err error)
	Snapshot() (snapshot []byte, err error)
	Restore(snapshot []byte) (err error)
}
