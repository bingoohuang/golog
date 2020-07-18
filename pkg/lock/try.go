package lock

type Try struct {
	C chan struct{}
}

func NewTry() *Try {
	t := &Try{
		C: make(chan struct{}, 1),
	}

	t.C <- struct{}{}

	return t
}

func (t *Try) TryLock() bool {
	select {
	case <-t.C:
		return true
	default:
		return false
	}
}

func (t *Try) Unlock() bool {
	select {
	case t.C <- struct{}{}:
		return true
	default:
		return false
	}
}
