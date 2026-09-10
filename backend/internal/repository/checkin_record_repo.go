package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

type checkinRepository struct {
	db *sql.DB
}

// NewCheckinRepository creates the SQL-backed check-in repository.
func NewCheckinRepository(db *sql.DB) service.CheckinRepository {
	return &checkinRepository{db: db}
}

// NewAdminCheckinRepository exposes the same SQL implementation through the
// admin-only query port without widening the user-facing repository interface.
func NewAdminCheckinRepository(db *sql.DB) service.AdminCheckinRepository {
	return &checkinRepository{db: db}
}

func (r *checkinRepository) GetUserState(ctx context.Context, userID int64) (*service.CheckinUserState, error) {
	var state service.CheckinUserState
	err := r.db.QueryRowContext(ctx, `
		SELECT u.role, u.status, u.balance, u.created_at,
		       (COALESCE(u.total_recharged, 0) > 0 OR EXISTS (
				SELECT 1 FROM redeem_codes rc
				WHERE rc.used_by = u.id AND rc.status = 'used' AND rc.value > 0
				  AND rc.type IN ('balance', 'admin_balance')
		       )),
		       EXISTS (
				SELECT 1 FROM signup_risk_accounts sra
				WHERE sra.user_id = u.id AND sra.grant_allowed = FALSE
		       )
		FROM users u
		WHERE u.id = $1 AND u.deleted_at IS NULL`, userID).Scan(&state.Role, &state.Status, &state.Balance, &state.CreatedAt, &state.HasRecharge, &state.SignupGrantRestricted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrCheckinUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *checkinRepository) GetByDate(ctx context.Context, userID int64, date string) (*service.CheckinRecord, error) {
	return scanCheckinRecord(r.db.QueryRowContext(ctx, checkinRecordQuery+` WHERE user_id = $1 AND checkin_date = $2`, userID, date))
}

func (r *checkinRepository) List(ctx context.Context, userID int64, page, pageSize int) ([]service.CheckinRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM checkin_records WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, checkinRecordQuery+` WHERE user_id = $1 ORDER BY checkin_date DESC, id DESC LIMIT $2 OFFSET $3`, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.CheckinRecord, 0, pageSize)
	for rows.Next() {
		item, err := scanCheckinRecordFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *checkinRepository) AdminOverview(ctx context.Context, date string) (*service.AdminCheckinOverview, error) {
	var result service.AdminCheckinOverview
	result.BusinessDate = date
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE mode = 'normal'),
		       COUNT(*) FILTER (WHERE mode = 'lucky'),
		       COALESCE(SUM(reward_amount) FILTER (WHERE reward_amount > 0), 0),
		       COALESCE(SUM(reward_amount) FILTER (WHERE reward_amount < 0), 0)
		FROM checkin_records WHERE checkin_date = $1`, date).Scan(
		&result.Total, &result.NormalCount, &result.LuckyCount, &result.PositiveTotal, &result.NegativeTotal)
	return &result, err
}

func (r *checkinRepository) AdminList(ctx context.Context, filter service.AdminCheckinRecordFilter) ([]service.AdminCheckinRecord, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 4)
	arg := func(value any) string { args = append(args, value); return fmt.Sprintf("$%d", len(args)) }
	if strings.TrimSpace(filter.Date) != "" {
		conditions = append(conditions, "r.checkin_date = "+arg(strings.TrimSpace(filter.Date)))
	}
	if filter.UserID != nil {
		conditions = append(conditions, "r.user_id = "+arg(*filter.UserID))
	}
	if strings.TrimSpace(filter.Email) != "" {
		conditions = append(conditions, "u.email ILIKE "+arg("%"+strings.TrimSpace(filter.Email)+"%"))
	}
	where := strings.Join(conditions, " AND ")
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM checkin_records r LEFT JOIN users u ON u.id = r.user_id WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitArg := arg(filter.PageSize)
	offsetArg := arg((filter.Page - 1) * filter.PageSize)
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.user_id, r.checkin_date, r.mode, r.reward_type, r.random_value, r.reward_amount,
		       r.balance_before, r.balance_after, r.checked_in_at, COALESCE(u.email, '')
		FROM checkin_records r LEFT JOIN users u ON u.id = r.user_id
		WHERE `+where+` ORDER BY r.checkin_date DESC, r.id DESC LIMIT `+limitArg+` OFFSET `+offsetArg, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.AdminCheckinRecord, 0, filter.PageSize)
	for rows.Next() {
		var item service.AdminCheckinRecord
		if err := rows.Scan(&item.ID, &item.UserID, &item.CheckinDate, &item.Mode, &item.RewardType, &item.RandomValue, &item.RewardAmount, &item.BalanceBefore, &item.BalanceAfter, &item.CheckedInAt, &item.UserEmail); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *checkinRepository) UpdateConfigIfVersion(ctx context.Context, expectedVersion int64, values map[string]string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	var rawVersion string
	err = tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1 FOR UPDATE`, service.SettingKeyCheckinConfigVersion).Scan(&rawVersion)
	if err != nil {
		return false, err
	}
	currentVersion, err := strconv.ParseInt(strings.TrimSpace(rawVersion), 10, 64)
	if err != nil || currentVersion < 1 {
		return false, service.ErrCheckinConfigInvalid
	}
	if currentVersion != expectedVersion {
		return false, nil
	}

	for key, value := range values {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, NOW())
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`, key, value); err != nil {
			return false, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE settings SET value = $1, updated_at = NOW() WHERE key = $2`,
		strconv.FormatInt(currentVersion+1, 10), service.SettingKeyCheckinConfigVersion); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (r *checkinRepository) Apply(ctx context.Context, userID int64, businessDate, mode string, calculate func(service.CheckinSettlementState) (decimal.Decimal, decimal.Decimal, string, error)) (*service.CheckinRecord, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	var role, status, balanceText string
	var checkinCount int64
	var hasRecharge bool
	if err := tx.QueryRowContext(ctx, `
		SELECT u.role,
		       u.status,
		       u.balance::text,
		       (SELECT COUNT(*) FROM checkin_records cr WHERE cr.user_id = u.id),
		       (COALESCE(u.total_recharged, 0) > 0 OR EXISTS (
				SELECT 1
				FROM redeem_codes rc
				WHERE rc.used_by = u.id
				  AND rc.status = 'used'
				  AND rc.value > 0
				  AND rc.type IN ('balance', 'admin_balance')
			))
		FROM users u
		WHERE u.id = $1 AND u.deleted_at IS NULL
		FOR UPDATE`, userID).Scan(&role, &status, &balanceText, &checkinCount, &hasRecharge); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, service.ErrCheckinUserNotFound
		}
		return nil, false, err
	}
	if status != service.StatusActive {
		return nil, false, service.ErrCheckinNotEligible
	}
	balance, err := decimal.NewFromString(balanceText)
	if err != nil {
		return nil, false, err
	}
	if balance.IsNegative() {
		return nil, false, service.ErrCheckinNegativeBalance
	}

	if record, err := scanCheckinRecord(tx.QueryRowContext(ctx, checkinRecordQuery+` WHERE user_id = $1 AND checkin_date = $2`, userID, businessDate)); err == nil {
		if err := tx.Commit(); err != nil {
			return nil, false, err
		}
		return record, false, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	reward, randomValue, rewardType, err := calculate(service.CheckinSettlementState{
		Balance: balance, CheckinCount: checkinCount, HasRecharge: hasRecharge,
	})
	if err != nil {
		return nil, false, err
	}
	var balanceAfterText string
	if err := tx.QueryRowContext(ctx, `
		UPDATE users
		SET balance = GREATEST(balance + $1, 0), updated_at = NOW()
		WHERE id = $2
		RETURNING balance::text`, reward.StringFixed(service.CheckinCalculationScale), userID).Scan(&balanceAfterText); err != nil {
		return nil, false, err
	}
	var record service.CheckinRecord
	err = tx.QueryRowContext(ctx, `
		INSERT INTO checkin_records
			(user_id, checkin_date, mode, reward_type, random_value, reward_amount, balance_before, balance_after, calculation_scale, checked_in_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, checkin_date, mode, reward_type, random_value, reward_amount, balance_before, balance_after, checked_in_at`,
		userID, businessDate, mode, rewardType,
		randomValue.StringFixed(service.CheckinCalculationScale), reward.StringFixed(service.CheckinCalculationScale),
		balanceText, balanceAfterText, service.CheckinCalculationScale).Scan(
		&record.ID, &record.CheckinDate, &record.Mode, &record.RewardType, &record.RandomValue, &record.RewardAmount,
		&record.BalanceBefore, &record.BalanceAfter, &record.CheckedInAt)
	if err != nil {
		return nil, false, err
	}
	record.UserID = userID
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, created_at, notes)
		VALUES ('SYS-CHECKIN-' || $1, $2, $3, $4, $5, $6, $6, $7)`,
		record.ID, service.RedeemTypeCheckin, reward.StringFixed(service.CheckinCalculationScale),
		service.StatusUsed, userID, record.CheckedInAt, service.CheckinHistoryModeNote(record.Mode)); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	return &record, true, nil
}

const checkinRecordQuery = `
	SELECT id, user_id, checkin_date, mode, reward_type, random_value, reward_amount,
	       balance_before, balance_after, checked_in_at
	FROM checkin_records`

type checkinRowScanner interface {
	Scan(dest ...any) error
}

func scanCheckinRecord(row checkinRowScanner) (*service.CheckinRecord, error) {
	var record service.CheckinRecord
	err := row.Scan(&record.ID, &record.UserID, &record.CheckinDate, &record.Mode, &record.RewardType, &record.RandomValue,
		&record.RewardAmount, &record.BalanceBefore, &record.BalanceAfter, &record.CheckedInAt)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func scanCheckinRecordFromRows(rows *sql.Rows) (*service.CheckinRecord, error) {
	return scanCheckinRecord(rows)
}
