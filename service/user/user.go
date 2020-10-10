package user

import (
	"context"

	"compound/core"

	"github.com/fox-one/mixin-sdk-go"
)

type userService struct {
	mixinClient *mixin.Client
}

// New new user service
func New(client *mixin.Client) core.IUserService {
	return &userService{mixinClient: client}
}

func (s *userService) Find(ctx context.Context, mixinID string) (*core.User, error) {
	profile, err := s.mixinClient.ReadUser(ctx, mixinID)
	if err != nil {
		return nil, err
	}

	user := core.User{
		MixinID: profile.UserID,
		Name:    profile.FullName,
		Avatar:  profile.AvatarURL,
	}

	return &user, nil
}

func (s *userService) Login(ctx context.Context, token string) (*core.User, error) {
	profile, err := mixin.UserMe(ctx, token)
	if err != nil {
		return nil, err
	}

	user := core.User{
		MixinID:     profile.UserID,
		Name:        profile.FullName,
		Avatar:      profile.AvatarURL,
		AccessToken: token,
	}

	return &user, nil
}
