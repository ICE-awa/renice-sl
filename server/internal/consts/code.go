package consts

const (
	CodeSuccess              = 0
	CodeBadRequest           = 400000
	CodeInvalidParam         = 400001
	CodeInvalidEmailCode     = 400002
	CodeInvalidIdentifier    = 400003
	CodeInvalidPassword      = 400004
	CodeInvalidBucket        = 400005
	CodeUnauthorized         = 401000
	CodeInvalidRefreshToken  = 401001
	CodeForbidden            = 403000
	CodeNotFound             = 404000
	CodeUserNotFound         = 404001
	CodeLinkNotFound         = 404002
	CodeConflict             = 409000
	CodeParamConflict        = 409001
	CodeTooManyRequest       = 429000
	CodeRateLimitExceeded    = 429001
	CodeInternalServerError  = 500000
	CodeFailedToGenerateCode = 500001
	CodeFailedToRedirect     = 500002
	CodeServiceUnavailable   = 503000
)
