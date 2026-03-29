package idgenerator

type IDGenerator interface {
	NewID() (int64, error)
}
