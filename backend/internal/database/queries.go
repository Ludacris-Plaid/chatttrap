package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct {
	Pool *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{Pool: pool}
}

func (q *Queries) ExecTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := q.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (q *Queries) GetProfile(ctx context.Context, userID string) (balance float64, isVIP bool, vipExpires *time.Time, err error) {
	err = q.Pool.QueryRow(ctx,
		`SELECT balance, is_vip, vip_expires_at FROM profiles WHERE id = $1`, userID).
		Scan(&balance, &isVIP, &vipExpires)
	return
}

func (q *Queries) AddBalance(ctx context.Context, userID string, amount float64) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE profiles SET balance = balance + $1 WHERE id = $2`, amount, userID)
	return err
}

func (q *Queries) DeductBalance(ctx context.Context, userID string, amount float64) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE profiles SET balance = GREATEST(balance - $1, 0) WHERE id = $2`, amount, userID)
	return err
}

func (q *Queries) InsertCall(ctx context.Context, id, userID, spoofedCID, spoofedName, destination, status string, tokensCost int, costUSD float64) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO calls (id, user_id, spoofed_caller_id, spoofed_name, destination_number, status, tokens_cost, cost_usd)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, userID, spoofedCID, spoofedName, destination, status, tokensCost, costUSD)
	return err
}

func (q *Queries) AddBalanceTx(tx pgx.Tx, ctx context.Context, userID string, amount float64) error {
	_, err := tx.Exec(ctx,
		`UPDATE profiles SET balance = balance + $1 WHERE id = $2`, amount, userID)
	return err
}

func (q *Queries) InsertTransactionTx(tx pgx.Tx, ctx context.Context, id, userID, ttype string, amount float64, tokens int, description, status string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO transactions (id, user_id, type, amount, tokens, description, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, userID, ttype, amount, tokens, description, status)
	return err
}

func (q *Queries) DeductBalanceTx(tx pgx.Tx, ctx context.Context, userID string, amount float64) error {
	_, err := tx.Exec(ctx,
		`UPDATE profiles SET balance = GREATEST(balance - $1, 0) WHERE id = $2`, amount, userID)
	return err
}

func (q *Queries) InsertCallTx(tx pgx.Tx, ctx context.Context, id, userID, spoofedCID, spoofedName, destination, status string, tokensCost int, costUSD float64) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO calls (id, user_id, spoofed_caller_id, spoofed_name, destination_number, status, tokens_cost, cost_usd)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, userID, spoofedCID, spoofedName, destination, status, tokensCost, costUSD)
	return err
}

func (q *Queries) UpdateCallEnded(ctx context.Context, callID string, duration, tokens int, cost float64) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE calls SET status = 'completed', duration_seconds = $1, tokens_cost = $2, cost_usd = $3, ended_at = NOW() WHERE id = $4`,
		duration, tokens, cost, callID)
	return err
}

func (q *Queries) InsertDTMF(ctx context.Context, id, callID, userID, digit string, timestampMs int) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO dtmf_logs (id, call_id, user_id, digit, timestamp_ms) VALUES ($1, $2, $3, $4, $5)`,
		id, callID, userID, digit, timestampMs)
	if err != nil {
		return err
	}
	_, err = q.Pool.Exec(ctx,
		`UPDATE calls SET dtmf_captured = COALESCE(dtmf_captured, '') || $1 WHERE id = $2`,
		digit, callID)
	return err
}

func (q *Queries) GetDTMFByCall(ctx context.Context, callID string) ([]struct{ Digit string; TimestampMs int }, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT digit, timestamp_ms FROM dtmf_logs WHERE call_id = $1 ORDER BY timestamp_ms ASC`, callID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []struct{ Digit string; TimestampMs int }
	for rows.Next() {
		var d struct{ Digit string; TimestampMs int }
		if err := rows.Scan(&d.Digit, &d.TimestampMs); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, nil
}

func (q *Queries) InsertScript(ctx context.Context, id, userID, targetName string, targetAge int, targetService, targetDetails, goal, scriptType, content string, tokensCost int) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO scripts (id, user_id, target_name, target_age, target_service, target_details, goal, script_type, content, tokens_cost)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		id, userID, targetName, targetAge, targetService, targetDetails, goal, scriptType, content, tokensCost)
	return err
}

func (q *Queries) InsertSMSLog(ctx context.Context, id, userID, phone, content, senderID, msgID string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO sms_logs (id, user_id, phone_number, content, sender_id, message_id, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'sent')`,
		id, userID, phone, content, senderID, msgID)
	return err
}

func (q *Queries) InsertCampaign(ctx context.Context, id, userID, senderID, content string, targets, sent int) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO sms_campaigns (id, user_id, sender_id, content, targets, sent_count, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'sent')`,
		id, userID, senderID, content, targets, sent)
	return err
}

func (q *Queries) InsertTransaction(ctx context.Context, id, userID, ttype string, amount float64, tokens int, description, status string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO transactions (id, user_id, type, amount, tokens, description, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, userID, ttype, amount, tokens, description, status)
	return err
}

func (q *Queries) UpgradeVIP(ctx context.Context, userID string, price float64, days int) error {
	_, err := q.Pool.Exec(ctx,
		fmt.Sprintf(`UPDATE profiles SET balance = balance - $1, is_vip = true, vip_expires_at = NOW() + INTERVAL '%d days' WHERE id = $2`, days), price, userID)
	return err
}

func (q *Queries) GetCalls(ctx context.Context, userID string, limit int) (pgx.Rows, error) {
	return q.Pool.Query(ctx,
		`SELECT id, spoofed_caller_id, spoofed_name, destination_number, status,
		        duration_seconds, tokens_cost, cost_usd, dtmf_captured, created_at
		 FROM calls WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
}

func (q *Queries) GetTransactions(ctx context.Context, userID string, limit int) (pgx.Rows, error) {
	return q.Pool.Query(ctx,
		`SELECT id, type, amount, tokens, status, description, created_at
		 FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
}

func (q *Queries) GetScripts(ctx context.Context, userID string) (pgx.Rows, error) {
	return q.Pool.Query(ctx,
		`SELECT id, title, target_name, target_service, goal, script_type, content, tokens_cost, is_library, created_at
		 FROM scripts WHERE user_id = $1 ORDER BY created_at DESC`, userID)
}

func (q *Queries) GetCall(ctx context.Context, callID string) (id, userID, scid, sname, dest, status, dtmf string, dur, tokens int, cost float64, createdAt, endedAt *time.Time, err error) {
	err = q.Pool.QueryRow(ctx,
		`SELECT id, user_id, spoofed_caller_id, spoofed_name, destination_number, status,
		        duration_seconds, tokens_cost, cost_usd, dtmf_captured, created_at, ended_at
		 FROM calls WHERE id = $1`, callID).
		Scan(&id, &userID, &scid, &sname, &dest, &status, &dur, &tokens, &cost, &dtmf, &createdAt, &endedAt)
	return
}
