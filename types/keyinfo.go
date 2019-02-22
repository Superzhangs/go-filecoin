package types

import (
	cbor "gx/ipfs/QmcZLyosDwMKdB6NLRsiss9HXzDPhVhhRtPy67JFKTDQDX/go-ipld-cbor"

	"github.com/filecoin-project/go-filecoin/address"
	"github.com/filecoin-project/go-filecoin/crypto"
)

func init() {
	cbor.RegisterCborType(KeyInfo{})
}

// KeyInfo is a key and its type used for signing
type KeyInfo struct {
	// Private key as bytes
	PrivateKey crypto.PrivateKey `json:"privateKey"`
	// Curve used to generate private key
	Curve string `json:"curve"`
}

// Unmarshal decodes raw cbor bytes into KeyInfo.
func (ki *KeyInfo) Unmarshal(b []byte) error {
	return cbor.DecodeInto(b, ki)
}

// Marshal KeyInfo into bytes.
func (ki *KeyInfo) Marshal() ([]byte, error) {
	return cbor.DumpObject(ki)
}

// Key returns the private key of KeyInfo
func (ki *KeyInfo) Key() *crypto.PrivateKey {
	return &ki.PrivateKey
}

// Type returns the type of curve used to generate the private key
func (ki *KeyInfo) Type() string {
	return ki.Curve
}

// Equals returns true if the KeyInfo is equal to other.
func (ki *KeyInfo) Equals(other *KeyInfo) bool {
	if ki == nil && other == nil {
		return true
	}
	if ki == nil || other == nil {
		return false
	}
	if ki.Curve != other.Curve {
		return false
	}

	return ki.PrivateKey.Equals(&other.PrivateKey)
}

// Address returns the address for this keyinfo
func (ki *KeyInfo) Address() (address.Address, error) {
	pub := ki.PublicKey()
	addrHash := address.Hash(pub.Serialize())

	// TODO: Use the address type we are running on from the config.
	return address.NewMainnet(addrHash), nil
}

// PublicKey returns the public key part as uncompressed bytes.
func (ki *KeyInfo) PublicKey() *crypto.PublicKey {
	return ki.Key().PublicKey()
}
