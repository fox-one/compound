package payee

import (
	"compound/core"
	"context"
)

func (w *Payee) handleUserAction(ctx context.Context, output *core.Output, actionType core.ActionType, followID string, body []byte) error {
	switch actionType {
	case core.ActionTypeSupply:
		return w.handleSupplyEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypeBorrow:
		return w.handleBorrowEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypeRedeem:
		return w.handleRedeemEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypeRepay:
		return w.handleRepayEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypePledge:
		return w.handlePledgeEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypeUnpledge:
		return w.handleUnpledgeEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypeQuickPledge:
		return w.handleQuickPledgeEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypeQuickBorrow:
		return w.handleQuickBorrowEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypeQuickRedeem:
		return w.handleQuickRedeemEvent(ctx, output, output.Sender, followID, body)
	case core.ActionTypeLiquidate:
		return w.handleLiquidationEvent(ctx, output, output.Sender, followID, body)
	default:
		return w.handleRefundEventV0(ctx, output, output.Sender, followID, core.ActionTypeRefundTransfer, core.ErrUnknown)
	}
}
