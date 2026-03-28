package ports

type IDGenerator interface {
	NewID() (int64, error)
}
