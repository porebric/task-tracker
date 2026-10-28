package api

import httpapi "github.com/porebric/healthstep-libs/api"

type ListTeamsResponse struct {
	httpapi.OK
	Teams []TeamMembership `json:"teams"`
}
