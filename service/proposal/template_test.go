package proposal

import (
	"compound/core"
	"fmt"
	"testing"

	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

func TestRenderProposal(t *testing.T) {
	p := &core.Proposal{
		TraceID: uuid.New(),
		Creator: uuid.New(),
		AssetID: uuid.New(),
		Amount:  decimal.New(7, 2),
		Action:  core.ActionTypeProposalAddMarket,
		Content: []byte(`{"a":"b"}`),
	}

	view := renderProposal(p)
	fmt.Println(string(view))
}
