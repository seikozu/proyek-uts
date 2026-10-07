package service

import (
	"proyek-uts/app/model"
	"strings"
)

func ApplyPatch(current model.Comic, req model.PatchComicRequest) model.Comic {
	if req.Title != nil {
		current.Title = strings.TrimSpace(*req.Title)
	}
	if req.Type != nil {
		current.Type = strings.TrimSpace(*req.Type)
	}
	if req.Author != nil {
		current.Author = strings.TrimSpace(*req.Author)
	}
	if req.Stock != nil {
		current.Stock = *req.Stock
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

func IsEmptyPatch(req model.PatchComicRequest) bool {
	return req.Title == nil && req.Type == nil && req.Author == nil && req.Stock == nil && req.IsActive == nil
}