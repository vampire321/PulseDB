//THIS REPO IS USED TO FETCH REQUIRED RECORD FROM A POSTGRESQL DATBASE. IT SPECIFICALLY TARGETS MONITORS THAT ARE DUE FOR CHECKING BASED ON THEIR LAST CHECKED TIME AND INTERVAL.
package postgres

import (
	"context"
)
type PostgresRepo struct {
    pool *pgxpool.Pool
}
//scheduler needs to find monitors that are due for checking
func (r *PostgresRepo) ListDueMonitors(ctx context.Context) ([]*Monitor, error) {
    rows, err := r.pool.Query(ctx,
        `SELECT id, name, url, interval_s, status, created_at, updated_at,
                last_checked_at
         FROM monitors
         WHERE status = 'active'
           AND (last_checked_at IS NULL
             OR last_checked_at + (interval_s * interval '1 second') <= NOW())
         ORDER BY last_checked_at ASC NULLS FIRST
         LIMIT 100`,
    )
	if err != nil {
        return nil, fmt.Errorf("list due monitors: %w", err)
    }
    defer rows.Close()

    var monitors []*Monitor
    for rows.Next() {
        m := &Monitor{}
        err := rows.Scan(&m.ID, &m.Name, &m.URL, &m.IntervalS,
            &m.Status, &m.CreatedAt, &m.UpdatedAt, &m.LastCheckedAt)
        if err != nil {
            return nil, fmt.Errorf("scan monitor: %w", err)
        }
        monitors = append(monitors, m)
    }
    return monitors, rows.Err()
}