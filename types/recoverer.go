package types

import "github.com/filecoin-project/go-filecoin/crypto"

// Recoverer is an interface for ecrecover
type Recoverer interface {
	Ecrecover(data []byte, sig Signature) (*crypto.PublicKey, error)
}
