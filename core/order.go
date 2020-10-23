package core

import "time"

type Order struct {
	ID            uint64    `json:"id"`
	SupplyAssetID string    `json:"supply_asset_id"`
	SupplyAmount  string    `json:"supply_amount"`
	BorrowAssetID string    `json:"borrow_asset_id"`
	BorrowAmount  string    `json:"borrow_amount"`
	UserID        string    `json:"user_id"`
	Status        int       `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}