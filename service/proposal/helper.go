package proposal

import (
	"context"

	"github.com/fox-one/pkg/uuid"
)

func (s *service) fetchAssetSymbol(ctx context.Context, assetID string) string {
	if uuid.IsNil(assetID) {
		return "ALL"
	}

	asset, err := s.client.ReadAsset(ctx, assetID)
	if err != nil {
		return assetID
	}

	return asset.Symbol
}

func (s *service) fetchUserName(ctx context.Context, userID string) string {
	user, err := s.client.ReadUser(ctx, userID)
	if err != nil {
		return userID
	}

	return user.FullName
}
