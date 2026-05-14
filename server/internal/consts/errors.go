package consts

import "errors"

var (
	ErrRegisterParamConflict = errors.New("username or email already exists")
	ErrInvalidEmailCode      = errors.New("invalid email code")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrFailedToGenerateCode  = errors.New("failed to generate code")
	ErrNoRowsAffected        = errors.New("no rows affected")
	ErrInvalidLink           = errors.New("invalid link")
)
