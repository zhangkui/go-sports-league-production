package repository

import (
	"context"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
)

// RoleRepo handles RBAC role/permission persistence.
type RoleRepo struct{ DB }

func NewRoleRepo(db DB) *RoleRepo { return &RoleRepo{db} }

func (r *RoleRepo) CreateRole(ctx context.Context, ro *models.Role, permIDs []int64) error {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO roles (code,name,description,is_builtin) VALUES (?,?,?,0)`,
		ro.Code, ro.Name, ro.Description)
	if err != nil {
		return rollback(tx, translateDup(err, "role code already exists"))
	}
	id, _ := res.LastInsertId()
	ro.ID = id
	for _, pid := range permIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO role_permissions (role_id,permission_id) VALUES (?,?)`, id, pid); err != nil {
			return rollback(tx, translateDup(err, "permission assignment failed"))
		}
	}
	return tx.Commit()
}

func (r *RoleRepo) UpdateRole(ctx context.Context, ro *models.Role, permIDs []int64, replacePerms bool) error {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE roles SET name=?,description=? WHERE id=?`, ro.Name, ro.Description, ro.ID); err != nil {
		return rollback(tx, err)
	}
	if replacePerms {
		if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, ro.ID); err != nil {
			return rollback(tx, err)
		}
		for _, pid := range permIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO role_permissions (role_id,permission_id) VALUES (?,?)`, ro.ID, pid); err != nil {
				return rollback(tx, translateDup(err, "permission assignment failed"))
			}
		}
	}
	return tx.Commit()
}

func (r *RoleRepo) ListRoles(ctx context.Context) ([]models.Role, error) {
	rows, err := r.QueryContext(ctx, `SELECT id,code,name,COALESCE(description,''),is_builtin,created_at,updated_at FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Role{}
	for rows.Next() {
		var ro models.Role
		var ib int
		if err := rows.Scan(&ro.ID, &ro.Code, &ro.Name, &ro.Description, &ib, &ro.CreatedAt, &ro.UpdatedAt); err != nil {
			return nil, err
		}
		ro.IsBuiltin = ib == 1
		out = append(out, ro)
	}
	return out, nil
}

func (r *RoleRepo) GetRole(ctx context.Context, id int64) (*models.Role, error) {
	var ro models.Role
	var ib int
	err := r.QueryRowContext(ctx, `SELECT id,code,name,COALESCE(description,''),is_builtin,created_at,updated_at FROM roles WHERE id=?`, id).
		Scan(&ro.ID, &ro.Code, &ro.Name, &ro.Description, &ib, &ro.CreatedAt, &ro.UpdatedAt)
	if err != nil {
		return nil, NotFound("role", id)
	}
	ro.IsBuiltin = ib == 1
	return &ro, nil
}

func (r *RoleRepo) DeleteRole(ctx context.Context, id int64) error {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, id); err != nil {
		return rollback(tx, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE role_id=?`, id); err != nil {
		return rollback(tx, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM roles WHERE id=? AND is_builtin=0`, id); err != nil {
		return rollback(tx, err)
	}
	return tx.Commit()
}

func (r *RoleRepo) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	rows, err := r.QueryContext(ctx, `SELECT id,resource,action,name FROM permissions ORDER BY resource,action`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Permission{}
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Resource, &p.Action, &p.Name); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *RoleRepo) RolePermissionCodes(ctx context.Context, roleID int64) ([]string, error) {
	rows, err := r.QueryContext(ctx, `SELECT p.resource,p.action FROM role_permissions rp JOIN permissions p ON p.id=rp.permission_id WHERE rp.role_id=?`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var res, act string
		if err := rows.Scan(&res, &act); err != nil {
			return nil, err
		}
		out = append(out, models.PermCode(res, act))
	}
	return out, nil
}

func (r *RoleRepo) CountRoleUsers(ctx context.Context, roleID int64) (int64, error) {
	var n int64
	err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_roles WHERE role_id=?`, roleID).Scan(&n)
	return n, err
}

var _ = fmt.Sprintf
