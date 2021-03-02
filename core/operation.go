package core

import "context"

// OperationScope operation scope
type OperationScope string

const (
	// OSLiquidation operation scope liquidation
	OSLiquidation OperationScope = "liquidation"
	// OSSupply operation scope supply
	OSSupply OperationScope = "supply"
	// OSPledge operation scope pledge
	OSPledge OperationScope = "pledge"
	// OSUnpledge operation scope unpledge
	OSUnpledge OperationScope = "unpledge"
	// OSRedeem operation scope redeem
	OSRedeem OperationScope = "redeem"
	// OSBorrow operation scope borrow
	OSBorrow OperationScope = "borrow"
	// OSRepay operation scope repay
	OSRepay OperationScope = "repay"
)

func (s OperationScope) String() string {
	return string(s)
}

// CheckScope check scope
func CheckScope(scope string) bool {
	return scope == string(OSLiquidation) ||
		scope == string(OSSupply) ||
		scope == string(OSPledge) ||
		scope == string(OSUnpledge) ||
		scope == string(OSRedeem) ||
		scope == string(OSBorrow) ||
		scope == string(OSRepay)
}

// AllowList allow list
type AllowList struct {
	ID     uint64         `sql:"PRIMARY_KEY;AUTO_INCREMENT" json:"id"`
	UserID string         `sql:"size:36;unique_index:idx_allowlists_user_scope" json:"user_id"`
	Scope  OperationScope `sql:"size:64;unique_index:idx_allowlists_user_scope" json:"scope"`
}

// IAllowListStore allow list store interface
type IAllowListStore interface {
	Create(ctx context.Context, allowList *AllowList) error
	Find(ctx context.Context, userID string, scope OperationScope) (*AllowList, error)
	Delete(ctx context.Context, userID string, scope OperationScope) error
}

// IAllowListService allow list service
type IAllowListService interface {
	AddAllowListScope(ctx context.Context, scope OperationScope) error
	RemoveAllowListScope(ctx context.Context, scope OperationScope) error
	AddAllowList(ctx context.Context, userID string, scope OperationScope) error
	RemoveAllowList(ctx context.Context, userID string, scope OperationScope) error
	IsScopeInAllowList(ctx context.Context, scope OperationScope) (bool, error)
	CheckAllowList(ctx context.Context, userID string, scope OperationScope) (bool, error)
}
