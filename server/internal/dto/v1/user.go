package v1

type UserRegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
}

type UserRegisterConflictResp struct {
	IsUsernameConflict bool `json:"is_username_conflict"`
	IsEmailConflict    bool `json:"is_email_conflict"`
}

type UserLoginReq struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required,min=8,max=72"`
}

type UserLoginResp struct {
	ExpiresIn int64 `json:"expires_in"`
}

type UserUpdateReq struct {
	ID       int64   `json:"id" db:"id"`
	Username *string `json:"username" db:"username"`
	Password *string `json:"password" db:"password"`
	Email    *string `json:"email" db:"email"`
	Role     *string `json:"role" db:"role"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
