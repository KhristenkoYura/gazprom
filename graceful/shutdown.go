package graceful

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Runner func(ctx context.Context) error

type Close func() error

func newCancel(fn func()) *cancel {
	c := &cancel{fn: fn}
	c.wg.Add(1)
	return c
}

type cancel struct {
	once sync.Once
	fn   func()
	wg   sync.WaitGroup
}

func (c *cancel) cancel() {
	c.once.Do(func() {
		c.fn()
		c.wg.Wait()
	})
}

func (c *cancel) done() {
	c.wg.Done()
}

type Shutdown struct {
	wg      sync.WaitGroup
	once    sync.Once
	cancels []*cancel
	err     chan error
}

func NewShutdown() *Shutdown {
	return &Shutdown{
		err: make(chan error),
		wg:  sync.WaitGroup{},
	}
}

func (s *Shutdown) Run(r Runner) {
	s.wg.Add(1)
	ctx, cancelRun := context.WithCancel(context.Background())

	c := newCancel(cancelRun)
	s.cancels = append(s.cancels, c)

	go func() {
		s.err <- r(ctx)
		c.done()

		s.Cancel()
		s.wg.Done()
	}()
}

func (s *Shutdown) Close(name string, c Close) {
	s.Run(func(ctx context.Context) error {
		<-ctx.Done()
		err := c()
		if err != nil {
			err = fmt.Errorf("%s: %w", name, err)
		}
		return err
	})
}

func (s *Shutdown) Cancel() {
	s.once.Do(func() {
		for i := len(s.cancels) - 1; i >= 0; i-- {
			s.cancels[i].cancel()
		}
	})
}

func (s *Shutdown) Wait(fn func(err error)) {
	go func() {
		s.wg.Wait()
		close(s.err)
	}()

	for err := range s.err {
		if err != nil {
			fn(err)
		}
	}
}

func Signal(ctx context.Context) (err error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case s := <-signals:
		err = fmt.Errorf("signal: %s", s)
	case <-ctx.Done():
	}

	return
}

func HTTPServer(srv *http.Server) Runner {
	return func(ctx context.Context) (err error) {
		errCh := make(chan error, 1)

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- fmt.Errorf("http.Server.ListenAndServe: %w", err)
			}
		}()

		select {
		case err = <-errCh:
		case <-ctx.Done():
			if err = srv.Shutdown(context.Background()); err != nil {
				err = fmt.Errorf("http.Server.Shutdown: %w", err)
			}
		}

		return err
	}
}
