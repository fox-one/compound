package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimSuffix(t *testing.T) {
	raws := []*RawOutput{
		{
			CreatedAt: 1,
		},
		{
			CreatedAt: 2,
		},
		{
			CreatedAt: 3,
		},
		{
			CreatedAt: 4,
		},
		{
			CreatedAt: 5,
		},
	}

	assert.Len(t, trimSuffix(raws), 4)

	raws[3].CreatedAt = 5
	assert.Len(t, trimSuffix(raws), 3)

	raws[2].CreatedAt = 5
	assert.Len(t, trimSuffix(raws), 2)

	raws[1].CreatedAt = 5
	assert.Len(t, trimSuffix(raws), 1)

	raws[0].CreatedAt = 5
	assert.Len(t, trimSuffix(raws), 5)
}
