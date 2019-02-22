package walletutil

import (
	"gx/ipfs/QmVmDhyTTUcQXFD1rRQ64fGLMSAoaQvNH3hwuaCFAPq2hy/errors"
	"gx/ipfs/QmZp3eKdYQHHAneECmeK6HhiMwTPufmjC8DuuaGKv3unvx/blake2b-simd"

	"github.com/filecoin-project/go-filecoin/crypto"
)

// Sign cryptographically signs `data` using the private key `priv`.
func Sign(priv *crypto.PrivateKey, data []byte) ([]byte, error) {
	hash := blake2b.Sum256(data)
	// sign the content
	sig, err := priv.Sign(hash[:])
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sign data")
	}

	return sig, nil
}

// Verify cryptographically verifies that 'sig' is the signed hash of 'data' with
// the public key `pk`.
func Verify(pk *crypto.PublicKey, data, signature []byte) (bool, error) {
	hash := blake2b.Sum256(data)
	// remove recovery id
	sig := signature[:len(signature)-1]
	return pk.Verify(hash[:], sig)
}

// Ecrecover returns an uncompressed public key that could produce the given
// signature from data.
// Note: The returned public key should not be used to verify `data` is valid
// since a public key may have N private key pairs
func Ecrecover(data, signature []byte) (*crypto.PublicKey, error) {
	hash := blake2b.Sum256(data)
	return crypto.EcRecover(hash[:], signature)
}
