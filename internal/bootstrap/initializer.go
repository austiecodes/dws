package bootstrap

// Initializer defines the interface for component initialization
type Initializer interface {
	LoadConfig() error
	Init() error
}

// InitializeAll initializes all components that implement the Initializer interface
func InitializeAll(initializers ...Initializer) error {
	for _, init := range initializers {
		if err := init.LoadConfig(); err != nil {
			return err
		}
		if err := init.Init(); err != nil {
			return err
		}
	}
	return nil
}
