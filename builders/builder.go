package builders

// RuntimeBuilder is the interface that defines how to build runtime environments
type RuntimeBuilder interface {
	Init() error
	PullImage() error
	SetupContainer() error
	Cleanup() error
}
