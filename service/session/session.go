package session

import (
	"context"
	"errors"

	"compound/core"

	"github.com/asaskevich/govalidator"
	"github.com/bluele/gcache"
	"github.com/dgrijalva/jwt-go"
	"github.com/fox-one/mixin-sdk-go"
	"golang.org/x/sync/singleflight"
)

// New new session
func New(users core.IUserStore, userz core.IUserService, capacity int, issuers []string) core.Session {
	var s core.Session = &session{
		users:   users,
		userz:   userz,
		issuers: issuers,
		sf:      &singleflight.Group{},
	}

	if capacity > 0 {
		s = &cacheSession{
			Session: s,
			tokens:  gcache.New(capacity).LRU().Build(),
		}
	}

	return s
}

type session struct {
	users   core.IUserStore
	userz   core.IUserService
	sf      *singleflight.Group
	issuers []string
}

func (s *session) Login(ctx context.Context, accessToken string) (*core.User, error) {
	user, err, _ := s.sf.Do(accessToken, func() (interface{}, error) {
		var claim struct {
			jwt.StandardClaims
			Scope string `json:"scp,omitempty"`
		}
		_, _ = jwt.ParseWithClaims(accessToken, &claim, nil)

		if claim.Scope != "FULL" && !govalidator.IsIn(claim.Issuer, s.issuers...) {
			return nil, errors.New("invalid issuer")
		}

		if jti := claim.Id; govalidator.IsUUID(jti) {
			ctx = mixin.WithRequestID(ctx, jti)
		}

		user, err := s.userz.Login(ctx, accessToken)
		if err != nil {
			return nil, err
		}

		if exists, err := s.users.Find(ctx, user.MixinID); err == nil {
			user = exists
		}

		return user, nil
	})

	if err != nil {
		return nil, err
	}

	return user.(*core.User), nil
}
