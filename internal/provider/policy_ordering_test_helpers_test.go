package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type memoryOrderingClient struct {
	mu       sync.Mutex
	ids      []string
	moves    int
	failMove int
	listErr  error
}

func (c *memoryOrderingClient) List(context.Context, string) ([]policyPosition, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.listErr != nil {
		return nil, c.listErr
	}
	result := make([]policyPosition, len(c.ids))
	for i, id := range c.ids {
		result[i] = policyPosition{ID: id, Order: int64(i + 1)}
	}
	return result, nil
}
func (c *memoryOrderingClient) Move(_ context.Context, _, id string, index int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.moves++
	if c.moves == c.failMove {
		return errors.New("simulated move failure")
	}
	found := -1
	for i, current := range c.ids {
		if current == id {
			found = i
			break
		}
	}
	if found < 0 || index < 0 || index >= int64(len(c.ids)) {
		return fmt.Errorf("invalid move")
	}
	c.ids = append(c.ids[:found], c.ids[found+1:]...)
	c.ids = append(c.ids, "")
	copy(c.ids[int(index)+1:], c.ids[int(index):])
	c.ids[index] = id
	return nil
}
func (c *memoryOrderingClient) set(ids ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ids = append([]string{}, ids...)
}
