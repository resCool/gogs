// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sync

import (
	"sync"
)

// SingleInstancePool is a pool of non-identical instances
// that only one instance with same identity is in the pool at a time.
// In other words, only instances with different identities can exist
// at the same time.
//
// This pool is particularly useful for performing tasks on same resource
// on the file system in different goroutines.
type SingleInstancePool struct {
	lock sync.Mutex

	// pool maintains locks for each instance in the pool.
	pool map[string]*sync.Mutex

	// count maintains the number of times an instance with same identity checks in
	// to the pool, and should be reduced to 0 (removed from map) by checking out
	// with same number of times.
	count map[string]int
}

// NewSingleInstancePool initializes and returns a new SingleInstancePool object.
func NewSingleInstancePool() *SingleInstancePool {
	return &SingleInstancePool{
		pool:  make(map[string]*sync.Mutex),
		count: make(map[string]int),
	}
}

// CheckIn checks in an instance to the pool and hangs while instance
// with same indentity is using the lock.
func (p *SingleInstancePool) CheckIn(identity string) {
	p.lock.Lock()

	lock, has := p.pool[identity]
	if !has {
		lock = &sync.Mutex{}
		p.pool[identity] = lock
	}
	p.count[identity]++

	p.lock.Unlock()
	lock.Lock()
}

// CheckOut checks out an instance from the pool and releases the lock
// to let other instances with same identity to grab the lock.
func (p *SingleInstancePool) CheckOut(identity string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.pool[identity].Unlock()
	if p.count[identity] == 1 {
		delete(p.pool, identity)
		delete(p.count, identity)
	} else {
		p.count[identity]--
	}
}
