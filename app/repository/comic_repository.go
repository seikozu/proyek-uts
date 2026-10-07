package repository

import (
	"context"
	"errors"
	"fmt"

	"proyek-uts/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type ComicRepository interface {
	FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Comic, error)
	FindByID(ctx context.Context, id int) (model.Comic, error)
	Create(ctx context.Context, c model.Comic) (model.Comic, error)
	Update(ctx context.Context, c model.Comic) (model.Comic, error)
	Delete(ctx context.Context, id int) error
}

type comicPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewComicRepository(pool *pgxpool.Pool) ComicRepository {
	return &comicPostgresRepository{pool: pool}
}

const comicColumns = "id, title, type, author, stock, is_active, owner_id, created_at"

func scanComic(row pgx.Row) (model.Comic, error) {
	var c model.Comic
	err := row.Scan(&c.ID, &c.Title, &c.Type, &c.Author, &c.Stock, &c.IsActive, &c.OwnerID, &c.CreatedAt)
	return c, err
}

func (r *comicPostgresRepository) FindAfterCursor(ctx context.Context, q model.CursorQuery) ([]model.Comic, error) {
	args := []any{}
	where := " WHERE 1 = 1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (title ILIKE $%d OR author ILIKE $%d)", len(args), len(args))
	}
	if q.Type != "" {
		args = append(args, q.Type)
		where += fmt.Sprintf(" AND type = $%d", len(args))
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}
	if q.After != nil {
		args = append(args, q.After.CreatedAt, q.After.ID)
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)-1, len(args))
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf("SELECT %s FROM comics%s ORDER BY created_at DESC, id DESC LIMIT $%d", comicColumns, where, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar komik: %w", err)
	}
	defer rows.Close()

	result := []model.Comic{}
	for rows.Next() {
		c, err := scanComic(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca row komik: %w", err)
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query komik: %w", err)
	}
	return result, nil
}

func (r *comicPostgresRepository) FindByID(ctx context.Context, id int) (model.Comic, error) {
	query := fmt.Sprintf("SELECT %s FROM comics WHERE id = $1", comicColumns)
	c, err := scanComic(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Comic{}, ErrNotFound
		}
		return model.Comic{}, fmt.Errorf("mengambil komik: %w", err)
	}
	return c, nil
}

func (r *comicPostgresRepository) Create(ctx context.Context, c model.Comic) (model.Comic, error) {
	query := `INSERT INTO comics (title, type, author, stock, is_active, owner_id)
			  VALUES ($1, $2, $3, $4, $5, $6)
			  RETURNING ` + comicColumns

	created, err := scanComic(r.pool.QueryRow(ctx, query, c.Title, c.Type, c.Author, c.Stock, c.IsActive, c.OwnerID))
	if err != nil {
		if isUniqueViolation(err) {
			return model.Comic{}, ErrDuplicate
		}
		return model.Comic{}, fmt.Errorf("menyimpan komik: %w", err)
	}
	return created, nil
}

func (r *comicPostgresRepository) Update(ctx context.Context, c model.Comic) (model.Comic, error) {
	query := `UPDATE comics SET title = $1, type = $2, author = $3, stock = $4, is_active = $5
			  WHERE id = $6
			  RETURNING ` + comicColumns

	updated, err := scanComic(r.pool.QueryRow(ctx, query, c.Title, c.Type, c.Author, c.Stock, c.IsActive, c.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Comic{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Comic{}, ErrDuplicate
		}
		return model.Comic{}, fmt.Errorf("memperbarui komik: %w", err)
	}
	return updated, nil
}

func (r *comicPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM comics WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("menghapus komik: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	Create(ctx context.Context, u model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

const userColumns = "id, username, email, password, role, is_active, created_at"

func scanUser(row pgx.Row) (model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	return u, err
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE LOWER(username) = LOWER($1)", userColumns)
	u, err := scanUser(r.pool.QueryRow(ctx, query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE id = $1", userColumns)
	u, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	query := `INSERT INTO users (username, email, password, role, is_active)
			  VALUES ($1, $2, $3, $4, $5)
			  RETURNING ` + userColumns

	created, err := scanUser(r.pool.QueryRow(ctx, query, u.Username, u.Email, u.Password, u.Role, u.IsActive))
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("mendaftarkan user: %w", err)
	}
	return created, nil
}

func (r *userPostgresRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	query := `UPDATE users SET role = $1 WHERE id = $2 RETURNING ` + userColumns
	updated, err := scanUser(r.pool.QueryRow(ctx, query, role, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengubah role user: %w", err)
	}
	return updated, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}