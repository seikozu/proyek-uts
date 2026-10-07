package model

import "time"

type Comic struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Author    string    `json:"author"`
	Stock     int       `json:"stock"`
	IsActive  bool      `json:"is_active"`
	OwnerID   *int      `json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateComicRequest struct {
	Title  string `json:"title"  validate:"required,min=3,max=150"`
	Type   string `json:"type"   validate:"required,oneof=manga manhwa manhua"`
	Author string `json:"author" validate:"required,min=3,max=100"`
	Stock  int    `json:"stock"  validate:"required,min=0"`
}

type ReplaceComicRequest struct {
	Title    string `json:"title"     validate:"required,min=3,max=150"`
	Type     string `json:"type"      validate:"required,oneof=manga manhwa manhua"`
	Author   string `json:"author"    validate:"required,min=3,max=100"`
	Stock    int    `json:"stock"     validate:"min=0"`
	IsActive bool   `json:"is_active"`
}

type PatchComicRequest struct {
	Title    *string `json:"title,omitempty"     validate:"omitnil,min=3,max=150"`
	Type     *string `json:"type,omitempty"      validate:"omitnil,oneof=manga manhwa manhua"`
	Author   *string `json:"author,omitempty"    validate:"omitnil,min=3,max=100"`
	Stock    *int    `json:"stock,omitempty"     validate:"omitnil,min=0"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type AssignRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorQuery struct {
	Limit    int
	After    *Cursor
	Search   string
	Type     string
	IsActive *bool
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}