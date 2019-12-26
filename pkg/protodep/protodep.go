package protodep

//go:generate bash ./generate.sh

const (
	DefaultDepDir = "vendor"
)

type DepFactory interface {
	Gather(opts *Config) error
}

type Manager struct {
	depFactories []DepFactory
}

func NewManager(cwd string) (*Manager, error) {
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

func (m *Manager) Gather(opts *Config) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	for _, v := range m.depFactories {
		if err := v.Gather(opts); err != nil {
			return err
		}
	}
	return nil
}
