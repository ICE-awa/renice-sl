package consts

import "errors"

var (
	ErrRegisterParamConflict = errors.New("username or email already exists")
	ErrInvalidEmailCode      = errors.New("invalid email code")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrFailedToGenerateCode  = errors.New("failed to generate code")
	ErrNoRowsAffected        = errors.New("no rows affected")
	ErrInvalidBucket         = errors.New("invalid bucket")
	ErrLinkNotFound          = errors.New("link not found")
	ErrLinkInactive          = errors.New("inactive link")
	ErrLinkExpired           = errors.New("expired link")
	ErrLinkPending           = errors.New("pending link")
	ErrLinkUnsafe            = errors.New("unsafe link")
	ErrLinkUnknown           = errors.New("unknown link")
	ErrURLNotAllowed         = errors.New("URL not allowed")
)
