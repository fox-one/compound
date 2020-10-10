package core

import "github.com/shopspring/decimal"

// Supply supply info
type Supply struct {
	// asset id
	UserID  string          `sql:"size:36" json:"user_id"`
	AssetID string          `sql:"size:36" json:"asset_id"`
	CTokens decimal.Decimal `sql:"type:decimal(24,8)" json:"c_tokens"`
}

// Redeem
// redeemAllowed