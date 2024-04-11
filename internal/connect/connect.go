package connect

import (
	"strings"

	"github.com/ShapleyIO/orion-api/api"
	"github.com/ShapleyIO/orion-api/internal/config"
)

type Services struct {
	handlers *api.Handlers
}

type closer interface {
	Close() error
}

type batchError struct {
	errs []error
}

func CreateServices(cfg *config.Config) (s *Services, err error) {
	s = &Services{}

	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	if s.handlers, err = api.NewHandlers(cfg); err != nil {
		// Log some error
	}
	return s, err
}

func (s *Services) Handlers() *api.Handlers {
	return s.handlers
}

func (s *Services) Close() error {
	objs := []closer{}
	return parallelClose(objs...)
}

func parallelClose(objs ...closer) error {
	closers := make([]func() error, 0, len(objs))
	for _, obj := range objs {
		closers = append(closers, obj.Close)
	}
	return parallel(closers...)
}

func parallel(fns ...func() error) error {
	errCh := make(chan error, len(fns))
	for _, fn := range fns {
		go func(fn func() error) { errCh <- fn() }(fn)
	}

	var errs []error
	for i := 0; i < len(fns); i++ {
		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return &batchError{errs}
	}

	return nil
}

func (e *batchError) Error() string {
	msgs := make([]string, 0, len(e.errs))
	for _, err := range e.errs {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}
