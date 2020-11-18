package core

import "strconv"

// ErrorCode int
type ErrorCode int

const (
	// ErrUnknown unkown
	ErrUnknown ErrorCode = 100000
	// ErrOperationForbidden operation forbidden
	ErrOperationForbidden ErrorCode = 100001

	// ErrMarketNotFound no market
	ErrMarketNotFound ErrorCode = 100100
	// ErrInvalidAmount invalid amount
	ErrInvalidAmount ErrorCode = 100101
	// ErrSupplyNotFound no supply
	ErrSupplyNotFound ErrorCode = 100102
	// ErrBorrowNotFound no borrow
	ErrBorrowNotFound ErrorCode = 100103
	// ErrInsufficientCollaterals insufficient collaterals
	ErrInsufficientCollaterals ErrorCode = 100104
	//ErrInsufficientLiquidity insufficient liquidity
	ErrInsufficientLiquidity ErrorCode = 100105
	// ErrRedeemNotAllowed redeeem not allowed
	ErrRedeemNotAllowed ErrorCode = 100106
	// ErrSeizeNotAllowed seize not allowed
	ErrSeizeNotAllowed ErrorCode = 100107
	// ErrInvalidPrice invalid price
	ErrInvalidPrice ErrorCode = 100108
	// ErrBorrowNotAllowed borrow not allowed
	ErrBorrowNotAllowed ErrorCode = 100109
)

func (e ErrorCode) String() string {
	return strconv.Itoa(int(e))
}

func (e ErrorCode) Error() string {
	return e.String()
}
