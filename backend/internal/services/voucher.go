package services

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5/pgxpool"
)

var voucherPrefixes = []string{"HUSH-10", "HUSH-25", "HUSH-50", "HUSH-100"}
var voucherValues = map[string]int{
	"HUSH-10":  20,
	"HUSH-25":  50,
	"HUSH-50":  100,
	"HUSH-100": 200,
}

type VoucherService struct {
	pool *pgxpool.Pool
}

func NewVoucherService() *VoucherService {
	return &VoucherService{}
}

func (s *VoucherService) SetPool(pool *pgxpool.Pool) {
	s.pool = pool
}

func (s *VoucherService) Generate(prefix string) (string, int, error) {
	tokens, ok := voucherValues[prefix]
	if !ok {
		tokens = 10
	}

	n, _ := crand.Int(crand.Reader, big.NewInt(999999))
	code := fmt.Sprintf("%s-%06d", prefix, n.Int64())
	return code, tokens, nil
}

func (s *VoucherService) Redeem(ctx context.Context, code string) (int, error) {
	if len(code) < 7 {
		return 0, fmt.Errorf("invalid voucher code")
	}
	prefix := code[:7]
	tokens, ok := voucherValues[prefix]
	if !ok {
		return 0, fmt.Errorf("unknown voucher type")
	}

	// Check DB if voucher exists and is unused
	if s.pool != nil {
		var isUsed bool
		err := s.pool.QueryRow(ctx,
			`SELECT is_used FROM vouchers WHERE code = $1`, code).Scan(&isUsed)
		if err == nil && isUsed {
			return 0, fmt.Errorf("voucher already redeemed")
		}
		// Mark as used (insert if not exists)
		s.pool.Exec(ctx,
			`INSERT INTO vouchers (code, tokens, is_used) VALUES ($1, $2, true)
			 ON CONFLICT (code) DO UPDATE SET is_used = true, used_at = NOW()`,
			code, tokens)
	}

	return tokens, nil
}
