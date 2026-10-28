package api

// Service error codes returned in the "code" field of error responses. Codes 1 to 8 are
// reserved by the HTTP layer (internal error, invalid request, unauthorized and so on).
const (
	CodeEmailTaken         = 100
	CodeInvalidCredentials = 101
	CodeUserNotFound       = 102
	CodeTeamNotFound       = 103
	CodeNotTeamMember      = 104
	CodeRoleNotAllowed     = 105
	CodeAlreadyMember      = 106
	CodeTaskNotFound       = 107
	CodeAssigneeNotMember  = 108
	CodeVersionConflict    = 109
	CodeFieldNotEditable   = 110
)
