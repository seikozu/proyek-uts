package service

import (
	"proyek-uts/app/model"
	"proyek-uts/helper"
	"strings"
)

func CanAccessComic(current model.AuthUser, comic model.Comic, perms *helper.PermissionSet, anyPermission string) bool {
	if comic.OwnerID != nil && current.UserID == *comic.OwnerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

func ValidateAssignRole(current model.AuthUser, targetID int, req model.AssignRoleRequest, perms *helper.PermissionSet) map[string]string {
	errs := map[string]string{}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		errs["role"] = "wajib diisi"
		return errs
	}

	if !perms.IsKnownRole(role) {
		errs["role"] = "role tidak dikenal, pilih salah satu dari: " + strings.Join(perms.KnownRoles(), ", ")
	}

	if current.UserID == targetID {
		errs["role"] = "tidak boleh mengubah role diri sendiri"
	}

	return errs
}