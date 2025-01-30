package pong

type UserError struct {
	info string
}

type InternalError struct {
	info string
}

func NewUserError(info string) error {
	return &UserError{info: info}
}

func NewInternalError(info string) error {
	return &InternalError{info: info}
}

func (u *UserError) Error() string {
	return u.info
}

func (i *InternalError) Error() string {
	return i.info
}
