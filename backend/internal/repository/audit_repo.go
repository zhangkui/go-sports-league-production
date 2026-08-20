package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/goxm2/sports-league/internal/models"
)

// AuditRepo handles audit log persistence.
type AuditRepo struct{ DB }

func NewAuditRepo(db DB) *AuditRepo { return &AuditRepo{db} }

func (r *AuditRepo) Create(ctx context.Context, tx *sql.Tx, a *models.AuditLog) error {
	q := `INSERT INTO audit_logs (user_id,username,action,resource,resource_id,method,path,status_code,ip,request_id,detail) VALUES (?,?,?,?,?,?,?,?,?,?,?)`
	res, err := execInTx(ctx, tx, r.DB, q, a.UserID, a.Username, a.Action, a.Resource, a.ResourceID, a.Method, a.Path, a.StatusCode, a.IP, a.RequestID, a.Detail)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	a.ID = id
	return nil
}

func (r *AuditRepo) List(ctx context.Context, userID int64, action, resource string, from, to time.Time, p models.Pagination) ([]models.AuditLog, int64, error) {
	where := "1=1"
	args := []any{}
	if userID > 0 {
		where += " AND user_id=?"
		args = append(args, userID)
	}
	if action != "" {
		where += " AND action=?"
		args = append(args, action)
	}
	if resource != "" {
		where += " AND resource=?"
		args = append(args, resource)
	}
	if !from.IsZero() {
		where += " AND created_at>=?"
		args = append(args, from)
	}
	if !to.IsZero() {
		where += " AND created_at<=?"
		args = append(args, to)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_logs WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := fmt.Sprintf(`SELECT id,user_id,COALESCE(username,''),action,resource,COALESCE(resource_id,''),COALESCE(method,''),COALESCE(path,''),status_code,COALESCE(ip,''),COALESCE(request_id,''),COALESCE(detail,''),created_at FROM audit_logs WHERE %s ORDER BY id DESC LIMIT %d OFFSET %d`, where, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.AuditLog{}
	for rows.Next() {
		var a models.AuditLog
		var uid sql.NullInt64
		var resID, method, path, ip, rid, detail, username sql.NullString
		var code sql.NullInt64
		if err := rows.Scan(&a.ID, &uid, &username, &a.Action, &a.Resource, &resID, &method, &path, &code, &ip, &rid, &detail, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		if uid.Valid {
			u := uid.Int64
			a.UserID = &u
		}
		a.Username = username.String
		a.ResourceID = resID.String
		a.Method = method.String
		a.Path = path.String
		if code.Valid {
			v := int(code.Int64)
			a.StatusCode = &v
		}
		a.IP = ip.String
		a.RequestID = rid.String
		a.Detail = detail.String
		out = append(out, a)
	}
	return out, total, nil
}

func (r *AuditRepo) GetByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	a := &models.AuditLog{}
	var uid sql.NullInt64
	var resID, method, path, ip, rid, detail, username sql.NullString
	var code sql.NullInt64
	err := r.QueryRowContext(ctx, `SELECT id,user_id,COALESCE(username,''),action,resource,COALESCE(resource_id,''),COALESCE(method,''),COALESCE(path,''),status_code,COALESCE(ip,''),COALESCE(request_id,''),COALESCE(detail,''),created_at FROM audit_logs WHERE id=?`, id).
		Scan(&a.ID, &uid, &username, &a.Action, &a.Resource, &resID, &method, &path, &code, &ip, &rid, &detail, &a.CreatedAt)
	if err != nil {
		return nil, NotFound("audit_log", id)
	}
	if uid.Valid {
		u := uid.Int64
		a.UserID = &u
	}
	a.Username = username.String
	a.ResourceID = resID.String
	a.Method = method.String
	a.Path = path.String
	if code.Valid {
		v := int(code.Int64)
		a.StatusCode = &v
	}
	a.IP = ip.String
	a.RequestID = rid.String
	a.Detail = detail.String
	return a, nil
}
