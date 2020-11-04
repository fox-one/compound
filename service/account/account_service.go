package account

import (
	"compound/core"
	"context"

	"github.com/shopspring/decimal"
)

type accountService struct {
	supplyStore core.ISupplyStore
	borrowStore core.IBorrowStore
}

// New new account service
func New() core.IAccountService {
	return nil
}

func (s *accountService) CalculateAccountLiquidity(ctx context.Context, userID string) (decimal.Decimal, error) {
	// supplies, e := s.supplyStore.FindByUser(ctx, userID)
	// if e != nil {
	// 	return decimal.Zero, e
	// }

	// borrows, e := s.borrowStore.FindByUser(ctx, userID)
	// if e != nil {
	// 	return decimal.Zero, e
	// }

	return decimal.Zero, nil
}
