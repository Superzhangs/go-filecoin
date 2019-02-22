package crypto_test

import (
	"testing"

	"gx/ipfs/QmPVkJMTeRC6iBByPWdrRkD3BE5UXsj5HPzb4kPqL186mS/testify/assert"

	"github.com/filecoin-project/go-filecoin/crypto"
)

func TestGenerateKey(t *testing.T) {
	assert := assert.New(t)

	key := crypto.GenerateKey()
	assert.Equal(len(key), 32)
	assert.NotEqual(key[0], 0)
}
