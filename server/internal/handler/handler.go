package handler

import v1 "github.com/ICE-awa/renice-sl/internal/handler/v1"

type Handlers struct {
	HealthH *HealthHandler
	AuthHV1 *v1.AuthHandler
	LinkHV1 *v1.LinkHandler
	StatHV1 *v1.StatHandler
}
