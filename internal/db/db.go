package db

import (
	"context"
	"encoding/json"
	"fmt"

	"yourmodule/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) SaveOrder(ctx context.Context, ord models.Order, rawJSON []byte) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// вставка с upsert по PK order_uid
	_, err = tx.Exec(ctx, `
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id,
                            delivery_service, shardkey, sm_id, date_created, oof_shard, payload)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
        ON CONFLICT (order_uid) DO UPDATE
        SET payload = EXCLUDED.payload,
            track_number = EXCLUDED.track_number,
            entry = EXCLUDED.entry,
            locale = EXCLUDED.locale,
            internal_signature = EXCLUDED.internal_signature,
            customer_id = EXCLUDED.customer_id,
            delivery_service = EXCLUDED.delivery_service,
            shardkey = EXCLUDED.shardkey,
            sm_id = EXCLUDED.sm_id,
            date_created = EXCLUDED.date_created,
            oof_shard = EXCLUDED.oof_shard,
            created_at = now()
    `, ord.OrderUID, ord.TrackNumber, ord.Entry, ord.Locale, ord.InternalSignature, ord.CustomerID,
		ord.DeliveryService, ord.ShardKey, ord.SmID, ord.DateCreated, ord.OofShard, rawJSON)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) GetOrder(ctx context.Context, orderUID string) (models.Order, []byte, error) {
	var raw json.RawMessage
	var o models.Order
	row := s.pool.QueryRow(ctx, `SELECT payload FROM orders WHERE order_uid = $1`, orderUID)
	err := row.Scan(&raw)
	if err != nil {
		return o, nil, err
	}
	if err := json.Unmarshal(raw, &o); err != nil {
		return o, raw, err
	}
	return o, raw, nil
}

func (s *Store) LoadAllOrders(ctx context.Context, limit int) (map[string]models.Order, error) {
	out := make(map[string]models.Order)
	q := `SELECT payload FROM orders ORDER BY created_at DESC`
	if limit > 0 {
		q = q + fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			continue
		}
		var o models.Order
		if err := json.Unmarshal(raw, &o); err != nil {
			continue
		}
		out[o.OrderUID] = o
	}
	return out, nil
}
