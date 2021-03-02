package operation

import (
	"compound/core"
	"context"
	"encoding/json"

	"github.com/fox-one/pkg/property"
	"github.com/jinzhu/gorm"
	"github.com/qinix/gods/lists/arraylist"
	"github.com/sirupsen/logrus"
)

const (
	// OperationKeyAllowListScope allow scope
	OperationKeyAllowListScope = "allow_list_scope"
)

type operationService struct {
	propertyStore  property.Store
	allowListStore core.IAllowListStore
}

// New new allow list service
func New(propertyStr property.Store, allowListStr core.IAllowListStore) core.IAllowListService {
	return &operationService{
		propertyStore:  propertyStr,
		allowListStore: allowListStr,
	}
}

func (s *operationService) AddAllowListScope(ctx context.Context, scope core.OperationScope) error {
	scopes, err := s.getAllowListScopes(ctx)
	if err != nil {
		return err
	}

	if !s.isScopeExists(ctx, scope, scopes) {
		return s.appendAllowListScope(ctx, scope, scopes)
	}

	return nil
}

func (s *operationService) RemoveAllowListScope(ctx context.Context, scope core.OperationScope) error {
	scopes, err := s.getAllowListScopes(ctx)
	if err != nil {
		return err
	}

	if s.isScopeExists(ctx, scope, scopes) {
		return s.removeAllowListScope(ctx, scope, scopes)
	}

	return nil
}

func (s *operationService) AddAllowList(ctx context.Context, userID string, scope core.OperationScope) error {
	al := core.AllowList{
		UserID: userID,
		Scope:  scope,
	}

	return s.allowListStore.Create(ctx, &al)
}

func (s *operationService) RemoveAllowList(ctx context.Context, userID string, scope core.OperationScope) error {
	_, err := s.allowListStore.Find(ctx, userID, scope)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return err
	}

	return s.allowListStore.Delete(ctx, userID, scope)
}

func (s *operationService) IsScopeInAllowList(ctx context.Context, scope core.OperationScope) (bool, error) {
	scopes, err := s.getAllowListScopes(ctx)
	if err != nil {
		return false, err
	}

	return s.isScopeExists(ctx, scope, scopes), nil
}

func (s *operationService) CheckAllowList(ctx context.Context, userID string, scope core.OperationScope) (bool, error) {
	_, err := s.allowListStore.Find(ctx, userID, scope)
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}

	return true, err
}

func (s *operationService) getAllowListScopes(ctx context.Context) (*arraylist.List, error) {
	v, err := s.propertyStore.Get(ctx, OperationKeyAllowListScope)
	if err != nil {
		return nil, err
	}
	scopeStr := v.String()

	logrus.Infoln("allow_list_scope:", scopeStr)

	var scopes []string
	err = json.Unmarshal([]byte(scopeStr), &scopes)
	if err != nil {
		return arraylist.New(), nil
	}

	l := arraylist.New()
	size := len(scopes)
	for i := 0; i < size; i++ {
		l.Add(scopes[i])
	}

	return l, nil
}

func (s *operationService) isScopeExists(ctx context.Context, scope core.OperationScope, scopes *arraylist.List) bool {
	return scopes.Contains(scope.String())
}

func (s *operationService) appendAllowListScope(ctx context.Context, scope core.OperationScope, scopes *arraylist.List) error {
	scopes.Add(scope.String())
	bs, err := scopes.ToJSON()
	if err != nil {
		return err
	}

	if err = s.propertyStore.Save(ctx, OperationKeyAllowListScope, string(bs)); err != nil {
		return err
	}

	return nil
}

func (s *operationService) removeAllowListScope(ctx context.Context, scope core.OperationScope, scopes *arraylist.List) error {
	scopes.Remove(scopes.IndexOf(scope.String()))
	bs, err := scopes.ToJSON()
	if err != nil {
		return err
	}

	if err = s.propertyStore.Save(ctx, OperationKeyAllowListScope, string(bs)); err != nil {
		return err
	}

	return nil
}
