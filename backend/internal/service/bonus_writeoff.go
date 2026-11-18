package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type BonusWriteoffPolicy struct {
	Enabled     bool      `json:"enabled"`
	EffectiveAt time.Time `json:"effective_at"`
	AmountLimit float64   `json:"amount_limit"`
	ConsentNote string    `json:"consent_note"`
}

type paymentBonusBatchContextKey struct{}

type BonusWriteoffRepository interface {
	GetBonusWriteoff(context.Context, int64, int) (json.RawMessage, error)
	SetBonusWriteoff(context.Context, int64, int64, BonusWriteoffPolicy) error
}

func (s *UserService) GetBonusWriteoff(ctx context.Context, id int64, page int) (json.RawMessage, error) {
	r, ok := s.userRepo.(BonusWriteoffRepository)
	if !ok {
		return nil, errors.New("bonus writeoff repository unavailable")
	}
	if page < 1 {
		page = 1
	}
	if page > 100000 {
		page = 100000
	}
	return r.GetBonusWriteoff(ctx, id, page)
}

func (s *UserService) SetBonusWriteoff(ctx context.Context, id, actor int64, p BonusWriteoffPolicy) error {
	p.ConsentNote = strings.TrimSpace(p.ConsentNote)
	if math.IsNaN(p.AmountLimit) || math.IsInf(p.AmountLimit, 0) || p.AmountLimit < 0 || p.AmountLimit > 1e9 || len(p.ConsentNote) > 2000 || (p.Enabled && p.ConsentNote == "") {
		return infraerrors.BadRequest("INVALID_WRITEOFF_POLICY", "Invalid writeoff policy or missing consent note")
	}
	if p.EffectiveAt.IsZero() {
		p.EffectiveAt = time.Now().UTC()
	}
	r, ok := s.userRepo.(BonusWriteoffRepository)
	if !ok {
		return errors.New("bonus writeoff repository unavailable")
	}
	return r.SetBonusWriteoff(ctx, id, actor, p)
}

// Invoked after crediting the payment redeem code, in the same transaction.
func createPaymentBonusBatch(ctx context.Context, client *dbent.Client, userID int64, code string) error {
	_, err := client.ExecContext(ctx, `SELECT register_recharge_bonus_batch($1,$2)`, userID, code)
	return err
}
