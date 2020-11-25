package mtg

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestScan(t *testing.T) {
	var (
		typ   int8       = 1
		uid              = newUUID()
		str              = "123"
		data  RawMessage = make([]byte, 100)
	)

	_, _ = io.ReadFull(rand.Reader, data)

	body, err := Encode(typ, uid, str, string(data))
	require.Nil(t, err)

	var (
		dtyp   int8
		duid   uuid.UUID
		dstr   string
		ddata  RawMessage
	)

	remain, err := Scan(body, &dtyp)
	require.Nil(t, err)
	assert.Equal(t, dtyp, typ)

	_, err = Scan(remain, &duid, &dstr, &ddata)
	require.Nil(t, err)

	assert.Equal(t, uid.String(), duid.String())
	assert.Equal(t, str, dstr)
	assert.Equal(t, data, ddata)
}
