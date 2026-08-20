package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/goxm2/sports-league/internal/models"
)

// VenueRepo handles venue + availability persistence.
type VenueRepo struct{ DB }

func NewVenueRepo(db DB) *VenueRepo { return &VenueRepo{db} }

const venueCols = `id, code, name, COALESCE(address,''), capacity, COALESCE(sport,''), status, created_at, updated_at`

func (r *VenueRepo) scanVenue(s interface{ Scan(...any) error }) (*models.Venue, error) {
	v := &models.Venue{}
	if err := s.Scan(&v.ID, &v.Code, &v.Name, &v.Address, &v.Capacity, &v.Sport, &v.Status, &v.CreatedAt, &v.UpdatedAt); err != nil {
		return nil, err
	}
	return v, nil
}

func (r *VenueRepo) Create(ctx context.Context, v *models.Venue) error {
	res, err := r.ExecContext(ctx, `INSERT INTO venues (code,name,address,capacity,sport,status) VALUES (?,?,?,?,?,?)`,
		v.Code, v.Name, v.Address, v.Capacity, v.Sport, v.Status)
	if err != nil {
		return translateDup(err, "venue code already exists")
	}
	id, _ := res.LastInsertId()
	v.ID = id
	return nil
}

func (r *VenueRepo) GetByID(ctx context.Context, id int64) (*models.Venue, error) {
	row := r.QueryRowContext(ctx, `SELECT `+venueCols+` FROM venues WHERE id=?`, id)
	v, err := r.scanVenue(row)
	if err != nil {
		return nil, NotFound("venue", id)
	}
	return v, nil
}

func (r *VenueRepo) List(ctx context.Context, status, sport string, p models.Pagination, orderBy string) ([]models.Venue, int64, error) {
	where := "1=1"
	args := []any{}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	if sport != "" {
		where += " AND sport=?"
		args = append(args, sport)
	}
	var total int64
	if err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM venues WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if orderBy == "" {
		orderBy = "id ASC"
	}
	q := fmt.Sprintf(`SELECT %s FROM venues WHERE %s ORDER BY %s LIMIT %d OFFSET %d`, venueCols, where, orderBy, p.Limit(), p.Offset())
	rows, err := r.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []models.Venue{}
	for rows.Next() {
		v, err := r.scanVenue(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *v)
	}
	return out, total, nil
}

func (r *VenueRepo) Update(ctx context.Context, v *models.Venue) error {
	_, err := r.ExecContext(ctx, `UPDATE venues SET name=?,address=?,capacity=?,sport=?,status=? WHERE id=?`,
		v.Name, v.Address, v.Capacity, v.Sport, v.Status, v.ID)
	return err
}

func (r *VenueRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.ExecContext(ctx, `DELETE FROM venues WHERE id=?`, id)
	return err
}

// Availability -----------------------------------------------------------

func (r *VenueRepo) ListAvailability(ctx context.Context, venueID int64) ([]models.VenueAvailability, error) {
	rows, err := r.QueryContext(ctx, `SELECT id,venue_id,weekday,start_time,end_time FROM venue_availability WHERE venue_id=? ORDER BY weekday,start_time`, venueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.VenueAvailability{}
	for rows.Next() {
		var va models.VenueAvailability
		if err := rows.Scan(&va.ID, &va.VenueID, &va.Weekday, &va.StartTime, &va.EndTime); err != nil {
			return nil, err
		}
		out = append(out, va)
	}
	return out, nil
}

func (r *VenueRepo) SetAvailability(ctx context.Context, venueID int64, slots []models.VenueAvailability) error {
	tx, err := r.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM venue_availability WHERE venue_id=?`, venueID); err != nil {
		return rollback(tx, err)
	}
	for _, s := range slots {
		if _, err := tx.ExecContext(ctx, `INSERT INTO venue_availability (venue_id,weekday,start_time,end_time) VALUES (?,?,?,?)`, venueID, s.Weekday, s.StartTime, s.EndTime); err != nil {
			return rollback(tx, translateDup(err, "duplicate availability slot"))
		}
	}
	return tx.Commit()
}

func (r *VenueRepo) IsVenueAvailable(ctx context.Context, venueID int64, weekday int, start, end string) (bool, error) {
	var n int
	err := r.QueryRowContext(ctx, `SELECT COUNT(*) FROM venue_availability WHERE venue_id=? AND weekday=? AND start_time<=? AND end_time>=?`, venueID, weekday, start, end).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

var _ = sql.ErrNoRows
