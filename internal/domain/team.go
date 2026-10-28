package domain

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func ParseRole(raw string) (Role, bool) {
	switch r := Role(raw); r {
	case RoleOwner, RoleAdmin, RoleMember:
		return r, true
	default:
		return "", false
	}
}

// CanManageTeam reports whether the role may invite members and read team analytics.
func (r Role) CanManageTeam() bool {
	return r == RoleOwner || r == RoleAdmin
}

type Team struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
}

func (Team) TableName() string { return "teams" }

type TeamMember struct {
	TeamID    int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"primaryKey"`
	Role      Role
	CreatedAt time.Time
}

func (TeamMember) TableName() string { return "team_members" }

// TeamMembership is a team joined with the role of a particular user in it.
type TeamMembership struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
	Role      Role
}
