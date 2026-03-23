package repo

import "errors"

var (
    ErrUserAlreadyExists         = errors.New("user already exists")
    ErrUserNotFound              = errors.New("user not found")
    ErrOrganizationAlreadyExists = errors.New("organization already exists")
)

