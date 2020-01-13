package skiplist

import (
	"sync"
	"time"
)

type uniqueTS struct {
	delta int64
	mu    sync.Mutex
}

func (u *uniqueTS) Now() (res int64) {
	u.mu.Lock()
	res = time.Now().UnixNano() + u.delta
	u.delta++
	if u.delta >= 512 {
		u.delta = 0
	}
	u.mu.Unlock()
	return
}

func newUniqueTS() *uniqueTS {
	return &uniqueTS{delta: 0}
}
