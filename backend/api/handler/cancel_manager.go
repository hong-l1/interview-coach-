package handler

import (
	"context"
	"sync"
)

type CancelRegistry struct {
	mu      sync.RWMutex
	cancels map[string]context.CancelFunc
}

func NewCancelRegistry() *CancelRegistry {
	return &CancelRegistry{
		cancels: make(map[string]context.CancelFunc),
	}
}

func (r *CancelRegistry) Set(sessionID string, cancel context.CancelFunc) {
	r.mu.Lock()
	r.cancels[sessionID] = cancel
	r.mu.Unlock()
}

func (r *CancelRegistry) Cancel(sessionID string) bool {
	r.mu.Lock()
	cancel := r.cancels[sessionID]
	delete(r.cancels, sessionID)
	r.mu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

func (r *CancelRegistry) Delete(sessionID string) {
	r.mu.Lock()
	delete(r.cancels, sessionID)
	r.mu.Unlock()
}
