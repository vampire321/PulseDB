//THIS REPO IS USED TO FETCH REQUIRED RECORD FROM A POSTGRESQL DATBASE. IT SPECIFICALLY TARGETS MONITORS THAT ARE DUE FOR CHECKING BASED ON THEIR LAST CHECKED TIME AND INTERVAL.
package monitor

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) Create(ctx context.Context, m *Monitor) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO monitors (name, url, interval_s)
		 VALUES ($1, $2, $3)
		 RETURNING id, status, created_at, updated_at`,
		m.Name, m.URL, m.IntervalS,
	).Scan(&m.ID, &m.Status, &m.CreatedAt, &m.UpdatedAt)
}

func (r *PostgresRepo) GetByID(ctx context.Context, id string) (*Monitor, error) {
	m := &Monitor{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, url, interval_s, status, created_at, updated_at,
		        last_checked_at
		 FROM monitors WHERE id = $1`, id,
	).Scan(&m.ID, &m.Name, &m.URL, &m.IntervalS, &m.Status,
		&m.CreatedAt, &m.UpdatedAt, &m.LastCheckedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get monitor: %w", err)
	}
	return m, nil
}

func (r *PostgresRepo) List(ctx context.Context) ([]*Monitor, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, url, interval_s, status, created_at, updated_at,
		        last_checked_at
		 FROM monitors ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list monitors: %w", err)
	}
	defer rows.Close()
	var monitors []*Monitor
	for rows.Next() {
		m := &Monitor{}
		if err := rows.Scan(&m.ID, &m.Name, &m.URL, &m.IntervalS,
			&m.Status, &m.CreatedAt, &m.UpdatedAt, &m.LastCheckedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}
func (r *PostgresRepo) Update(ctx context.Context, m *Monitor) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE monitors
		 SET name=$1, url=$2, interval_s=$3, updated_at=NOW(), last_checked_at=$4
		 WHERE id=$5`,
		m.Name, m.URL, m.IntervalS, m.LastCheckedAt, m.ID,
	)
	if err != nil {
		return fmt.Errorf("update monitor: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *PostgresRepo) Delete(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM monitors WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *PostgresRepo) SaveCheck(ctx context.Context, check *Check) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO checks (monitor_id, status_code, response_ms, error)
		 VALUES ($1, $2, $3, $4)`,
		check.MonitorID, check.StatusCode, check.ResponseMs, check.Error,
	)
	return err
}

func (r *PostgresRepo) ListChecks(ctx context.Context, monitorID string) ([]*Check, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, monitor_id, checked_at, status_code, response_ms, error
		 FROM checks WHERE monitor_id = $1
		 ORDER BY checked_at DESC LIMIT 50`, monitorID)
	if err != nil {
		return nil, fmt.Errorf("list checks: %w", err)
	}
	defer rows.Close()
	var checks []*Check
	for rows.Next() {
		c := &Check{}
		if err := rows.Scan(&c.ID, &c.MonitorID, &c.CheckedAt,
			&c.StatusCode, &c.ResponseMs, &c.Error); err != nil {
			return nil, fmt.Errorf("scan check: %w", err)
		}
		checks = append(checks, c)
	}
	return checks, rows.Err()
}
func (r *PostgresRepo) ListDueMonitors(ctx context.Context) ([]*Monitor, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, url, interval_s, status, created_at, updated_at,
		        last_checked_at
		 FROM monitors
		 WHERE status = 'active'
		   AND (last_checked_at IS NULL
		     OR last_checked_at + (interval_s * interval '1 second') <= NOW())
		 ORDER BY last_checked_at ASC NULLS FIRST LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("list due: %w", err)
	}
	defer rows.Close()
	var monitors []*Monitor
	for rows.Next() {
		m := &Monitor{}
		if err := rows.Scan(&m.ID, &m.Name, &m.URL, &m.IntervalS,
			&m.Status, &m.CreatedAt, &m.UpdatedAt, &m.LastCheckedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}