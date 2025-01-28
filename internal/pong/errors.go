package pong

type PongError struct {
	code int
	info string
}

type ErrType int

const (
	LogicError    ErrType = 100
	InternalError ErrType = 500
)
