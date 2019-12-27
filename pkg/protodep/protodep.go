package protodep

import (
	"context"
)

//go:generate bash ./generate.sh

const (
	DefaultDepDir = "vendor"
)

type DepFactory interface {
	Ensure(ctx context.Context, opts *Config) error
}

type Manager struct {
	depFactories []DepFactory
}

func NewManager(ctx context.Context, cwd string) (*Manager, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	goMod, err := NewGoModFactory(cwd)
	if err != nil {
		return nil, err
	}
	return &Manager{
		depFactories: []DepFactory{
			goMod,
		},
	}, nil
}

func (m *Manager) Ensure(ctx context.Context, opts *Config) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	for _, v := range m.depFactories {
		if err := v.Ensure(ctx, opts); err != nil {
			return err
		}
	}
	return nil
}
