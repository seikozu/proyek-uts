package service

import (
	"errors"

	"proyek-uts/app/model"
	"proyek-uts/app/repository"
	"proyek-uts/helper"

	"github.com/gofiber/fiber/v2"
)

type ComicService struct {
	repo  repository.ComicRepository
	perms *helper.PermissionSet
}

func NewComicService(repo repository.ComicRepository, perms *helper.PermissionSet) *ComicService {
	return &ComicService{repo: repo, perms: perms}
}

func translateError(err error, entity string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(entity + " tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("judul komik sudah terdaftar")
	default:
		return helper.Internal(err)
	}
}

func (s *ComicService) List(c *fiber.Ctx) error {
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseCursorQuery(c)
	if err != nil {
		return err
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	rows, err := s.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	if format == helper.FormatCSV {
		return helper.WriteComicsCSV(c, rows)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}

	return helper.SuccessCursor(c, "daftar komik berhasil diambil", rows, meta)
}

func (s *ComicService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	comic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "komik")
	}

	if !CanAccessComic(current, comic, s.perms, "comic:read:any") {
		return helper.Forbidden("tidak berhak mengakses data komik ini")
	}

	return helper.Success(c, fiber.StatusOK, "komik ditemukan", comic)
}

func (s *ComicService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateComicRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	ownerID := current.UserID
	created, err := s.repo.Create(ctx, model.Comic{
		Title:    req.Title,
		Type:     req.Type,
		Author:   req.Author,
		Stock:    req.Stock,
		IsActive: true,
		OwnerID:  &ownerID,
	})

	if err != nil {
		return translateError(err, "komik")
	}

	return helper.Success(c, fiber.StatusCreated, "komik berhasil ditambahkan", created)
}

func (s *ComicService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	comic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "komik")
	}

	if !CanAccessComic(current, comic, s.perms, "comic:update:any") {
		return helper.Forbidden("tidak berhak mengubah data komik ini")
	}

	var req model.ReplaceComicRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	comic.Title = req.Title
	comic.Type = req.Type
	comic.Author = req.Author
	comic.Stock = req.Stock
	comic.IsActive = req.IsActive

	updated, err := s.repo.Update(ctx, comic)
	if err != nil {
		return translateError(err, "komik")
	}

	return helper.Success(c, fiber.StatusOK, "komik berhasil diperbarui seluruhnya", updated)
}

func (s *ComicService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	comic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "komik")
	}

	if !CanAccessComic(current, comic, s.perms, "comic:update:any") {
		return helper.Forbidden("tidak berhak mengubah data komik ini")
	}

	var req model.PatchComicRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	patchedComic := ApplyPatch(comic, req)
	updated, err := s.repo.Update(ctx, patchedComic)
	if err != nil {
		return translateError(err, "komik")
	}

	return helper.Success(c, fiber.StatusOK, "komik berhasil diperbarui sebagian", updated)
}

func (s *ComicService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	comic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "komik")
	}

	if !CanAccessComic(current, comic, s.perms, "comic:delete") {
		return helper.Forbidden("tidak berhak menghapus komik ini")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(err, "komik")
	}

	return helper.NoContent(c)
}