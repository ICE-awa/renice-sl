package v1

import "time"

type ClickStatItem struct {
	Time  time.Time `json:"time"`
	Count int64     `json:"count"`
}

type UserStatItem struct {
	Time  time.Time `json:"time"`
	Count int64     `json:"count"`
}

type LinkStatItem struct {
	Time  time.Time `json:"time"`
	Count int64     `json:"count"`
}

type GetClickStatReq struct {
	Range  int    `form:"range" json:"range" binding:"required,min=1,max=168"`
	Bucket string `form:"bucket" json:"bucket"`
}

type GetUserStatReq struct {
	Range  int    `form:"range" json:"range" binding:"required,min=1,max=168"`
	Bucket string `form:"bucket" json:"bucket"`
}

type GetLinkStatReq struct {
	Range  int    `form:"range" json:"range" binding:"required,min=1,max=168"`
	Bucket string `form:"bucket" json:"bucket"`
}
