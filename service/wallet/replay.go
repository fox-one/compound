package wallet

import (
	"compound/core"
	"context"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
)

func Replay(walletz core.WalletService) core.WalletService {
	return &replayMode{walletz}
}

// replay mode 不会 pull 新的 outputs，也只消费已经是 spent 状态的 utxo
type replayMode struct {
	core.WalletService
}

func (s *replayMode) Spend(ctx context.Context, outputs []*core.Output, transfer *core.Transfer) (*core.RawTransaction, error) {
	const state = mixin.UTXOStateSpent

	for _, output := range outputs {
		if output.State != state {
			return nil, fmt.Errorf("state %q not allowed, must %q", output.State, state)
		}
	}

	return s.WalletService.Spend(ctx, outputs, transfer)
}
