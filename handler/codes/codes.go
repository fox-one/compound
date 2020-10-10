package codes

import (
	"strconv"

	"github.com/twitchtv/twirp"
)

const (
	CustomCodeKey = "custom_code"

	InvalidArguments = 10002

	InsufficientLiquiditySwapped = 20001
	InsufficientFundsSwapped     = 20002
	SlippageOutSwapped           = 20003
	SlippageOutDeposit           = 20004
	InsufficientFundsDeposit     = 20005
)

func With(err error, code int) error {
	twerr, ok := err.(twirp.Error)
	if !ok {
		twerr = twirp.InternalErrorWith(err)
	}

	return twerr.WithMeta(CustomCodeKey, strconv.Itoa(code))
}

func Get(code twirp.ErrorCode) int {
	switch code {
	case twirp.InvalidArgument:
		return InvalidArguments
	default:
		return twirp.ServerHTTPStatusFromErrorCode(code)
	}
}
