package metcap

import (
	"sync"
)

// ----- Flag ------
type Flag struct {
	*sync.Mutex
	val bool
}

func (f *Flag) Get() bool {
	f.Lock()
	defer f.Unlock()
	return f.val
}

func (f *Flag) Raise() {
	f.Lock()
	defer f.Unlock()
	f.val = true
}

func (f *Flag) Lower() {
	f.Lock()
	defer f.Unlock()
	f.val = false
}

func (f *Flag) Flip() {
	f.Lock()
	defer f.Unlock()
	f.val = !f.val
}
