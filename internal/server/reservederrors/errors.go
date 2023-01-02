package reservederrors

import "errors"

var ErrUserAlreadyExist = errors.New("user already exist")
var ErrUserNotFound = errors.New("user not found")

var ErrTokenNotFound = errors.New("token not found")
var ErrTokenInvalid = errors.New("invalid token")
