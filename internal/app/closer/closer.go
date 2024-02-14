package closer

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Closer struct {
	mu    sync.Mutex
	funcs []Func
	names []string
}

func (c *Closer) Add(n string, f Func) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.names = append(c.names, n)
	c.funcs = append(c.funcs, f)
}

func (c *Closer) Close(ctx context.Context) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		complete = make(chan struct{}, 1)
	)

	go func() {
		for i, f := range c.funcs {
			if errF := f(ctx); errF != nil {
				err = errors.Join(err, fmt.Errorf("close %s error: %w", c.names[i], errF))
			}
		}
		complete <- struct{}{}
	}()

	select {
	case <-complete:
		break
	case <-ctx.Done():
		return fmt.Errorf("shutdown cancelled: %s", ctx.Err())
	}

	return
}

type Func func(ctx context.Context) error
