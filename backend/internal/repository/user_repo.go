package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
)

// UserRepo handles user persistence.
type UserRepo struct{ DB }

func NewUserRepo(db DB) *UserRepo { return &UserRepo{db} }

const userCols = `id, username, email, password_hash, COALESCE(full_name,''), COALESCE(phone,''), status, last_login_at, created_at, updated_at`

func (r *UserRepo) scanUser(s interface{ Scan(...any) error }) (*models.User, error) {
	u := &models.User{}
	var lastLogin sql.NullTime
	if err := s.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.Status, &lastLogin, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	if lastLogin.Valid {
		u.LastLoginAt = &lastLogin.Time
	}
	return u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	res, err := r.ExecContext(ctx, `INSERT INTO users (username,email,password_hash,full_name,phone,status) VALUES (?,?,?,?,?,1)`,
		u.Username, u.Email, u.PasswordHash, u.FullName, u.Phone)
	if err != nil {
		return translateDup(err, "username or email already exists")
	}
	id, _ := res.LastInsertId()
	u.ID = id
	u.Status = 1
	return nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	row := r.QueryRowContext(ctx, `SELECT `+userCols+` FROM users WHERE username=? LIMIT 1`, username)
	u, err := r.scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("user lookup: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.QueryRowContext(ctx, `SELECT `+userCols+` FROM users WHERE email=? LIMIT 1`, email)
	u, err := r.scanUser(row)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	row := r.QueryRowContext(ctx, `SELECT `+userCols+` FROM users WHERE id=? LIMIT 1`, id)
	u, err := r.scanUser(row)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) List(ctx context.Context, keyword string, p models.Pagination, orderBy string) ([]models.User, int64, error) {
	where := "1=1"
	args := []any{}
	if keyword != "" {
		where = "(username LIKE ? OR email LIKE ? OR full_name LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if orderBy == "" {
		orderBy = "id DESC"
	}
	q := fmt.Sprintf(`SELECT %s FROM users WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, userCols, where, orderBy, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.User{}
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *u)
	}
	return out, total, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, hash string) error {
	_, err := r.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, hash, id)
	return err
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id int64, fullName, phone, email *string, status *int) error {
	set := []string{}
	args := []any{}
	if fullName != nil {
		set = append(set, "full_name=?")
		args = append(args, *fullName)
	}
	if phone != nil {
		set = append(set, "phone=?")
		args = append(args, *phone)
	}
	if email != nil {
		set = append(set, "email=?")
		args = append(args, *email)
	}
	if status != nil {
		set = append(set, "status=?")
		args = append(args, *status)
	}
	if len(set) == 0 {
		return nil
	}
	args = append(args, id)
	q := "UPDATE users SET " + joinStrings(set, ", ") + " WHERE id=?"
	_, err := r.ExecContext(ctx, q, args...)
	return err
}

func (r *UserRepo) TouchLogin(ctx context.Context, id int64) error {
	_, err := r.ExecContext(ctx, `UPDATE users SET last_login_at=NOW(3) WHERE id=?`, id)
	return err
}

func (r *UserRepo) LoadRolesAndPerms(ctx context.Context, u *models.User) error {
	rows, err := r.QueryContext(ctx, `
		SELECT r.id, r.code, r.name, COALESCE(r.description,''), r.is_builtin
		FROM user_roles ur JOIN roles r ON r.id=ur.role_id
		WHERE ur.user_id=? ORDER BY r.id`, u.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	roles := []models.Role{}
	roleIDs := []any{}
	for rows.Next() {
		var ro models.Role
		var ib int
		if err := rows.Scan(&ro.ID, &ro.Code, &ro.Name, &ro.Description, &ib); err != nil {
			return err
		}
		ro.IsBuiltin = ib == 1
		roles = append(roles, ro)
		roleIDs = append(roleIDs, ro.ID)
	}
	u.Roles = roles
	if len(roleIDs) == 0 {
		u.Permissions = []string{}
		return nil
	}
	ph := stringsRepeat("?", len(roleIDs))
	rows2, err := r.QueryContext(ctx, `
		SELECT DISTINCT p.resource, p.action
		FROM role_permissions rp JOIN permissions p ON p.id=rp.permission_id
		WHERE rp.role_id IN (`+ph+`)`, roleIDs...)
	if err != nil {
		return err
	}
	defer rows2.Close()
	perms := []string{}
	for rows2.Next() {
		var res, act string
		if err := rows2.Scan(&res, &act); err != nil {
			return err
		}
		perms = append(perms, models.PermCode(res, act))
	}
	u.Permissions = perms
	return nil
}

func (r *UserRepo) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=?`, userID); err != nil {
		return rollback(tx, err)
	}
	for _, rid := range roleIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO user_roles (user_id,role_id) VALUES (?,?)`, userID, rid); err != nil {
			return rollback(tx, translateDup(err, "role assignment failed"))
		}
	}
	return tx.Commit()
}

func joinStrings(parts []string, sep string) string {
	out := ""
	for i, s := range parts {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

func stringsRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}
