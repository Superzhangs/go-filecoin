package node

import (
	"context"
	"github.com/filecoin-project/go-filecoin/address"
	"github.com/filecoin-project/go-filecoin/protocol/storage"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePrivateKey(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	// should fail if less than 1024
	badKey, err := makePrivateKey(10)
	assert.Error(err, ErrLittleBits)
	assert.Nil(badKey)

	// 1024 should work
	okKey, err := makePrivateKey(1024)
	assert.NoError(err)
	assert.NotNil(okKey)

	// large values should work
	goodKey, err := makePrivateKey(4096)
	assert.NoError(err)
	assert.NotNil(goodKey)
}

func TestNode_getMinerOwnerPubKey(t *testing.T) {
	ctx := context.Background()
	seed := MakeChainSeed(t, TestGenCfg)
	configOpts := []ConfigOpt{RewarderConfigOption(&ZeroRewarder{})}
	tnode := MakeNodeWithChainSeed(t, seed, configOpts,
		PeerKeyOpt(PeerKeys[0]),
		AutoSealIntervalSecondsOpt(1),
	)
	seed.GiveKey(t, tnode, 0)
	mineraddr, minerOwnerAddr := seed.GiveMiner(t, tnode, 0)
	_, err := storage.NewMiner(ctx, mineraddr, minerOwnerAddr, tnode, tnode.Repo.DealsDatastore(), tnode.PorcelainAPI)
	assert.NoError(t, err)

	// it hasn't yet been saved to the MinerConfig; simulates incomplete CreateMiner, or no miner for the node
	pkey, err := tnode.getMinerActorPubKey()
	assert.NoError(t, err)
	assert.Nil(t, pkey)

	err = tnode.saveMinerConfig(minerOwnerAddr, address.Address{})
	assert.NoError(t, err)

	pkey, err = tnode.getMinerActorPubKey()
	assert.NoError(t, err)
	assert.NotNil(t, pkey)
}
