package compound

import (
	"context"
	"testing"
)

func TestCurrentBlock(t *testing.T) {
	currentBlock, e := CurrentBlock(context.Background(), 15, 1603366002)
	if e != nil {
		t.Error(e)
	}

	t.Log("currentBlock:", currentBlock)
}
