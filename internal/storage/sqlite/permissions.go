package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"SSO/internal/domain/models"
	"SSO/internal/storage"

	"github.com/mattn/go-sqlite3"
)

func (s *Storage) SavePermession(ctx context.Context, permission string) (int64, error) {
	const op = "storage.sqlite.SavePermession"

	stmt, err := s.db.Prepare("INSERT INTO permissions(permission) VALUES(?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, permission)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrPermExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) PermissionByID(ctx context.Context, id int64) (models.Permission, error) {
	const op = "storage.sqlite.PermessionByID"

	stmt, err := s.db.Prepare("SELECT permission FROM permissions WHERE id = ?")

	if err != nil {
		return models.Permission{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, id)

	permission := models.Permission{}

	err = row.Scan(&permission.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Permission{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.Permission{}, fmt.Errorf("%s: %w", op, err)
	}

	permission.ID = id
	return permission, nil
}

func (s *Storage) PermissionByName(ctx context.Context, name string) (models.Permission, error) {
	const op = "storage.sqlite.PermessionByName"

	stmt, err := s.db.Prepare("SELECT id FROM permissions WHERE permission = ?")

	if err != nil {
		return models.Permission{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, name)

	permission := models.Permission{}

	err = row.Scan(&permission.ID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Permission{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.Permission{}, fmt.Errorf("%s: %w", op, err)
	}

	permission.Name = name
	return permission, nil
}

func (s *Storage) UserPermissionsForApp(ctx context.Context, userID int64, appID int64) (map[int64](bool), error) {

	const op = "storage.sqlite.UserPermissions"

	stmt, err := s.db.Prepare("SELECT permissions.id from permissions inner JOIN users_permissions ON users_id = ? WHERE app_id = ?")

	if err != nil {
		return make(map[int64]bool), fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userID, appID)

	var resultPermissions map[int64]bool

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return make(map[int64]bool), fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
			}

			return make(map[int64]bool), fmt.Errorf("%s: %w", op, err)
		}
		resultPermissions[id] = true
	}

	return resultPermissions, nil
}
