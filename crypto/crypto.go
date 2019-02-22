package crypto

import (
	"io"
	"unsafe"
)

// #cgo LDFLAGS: -L${SRCDIR}/lib -lfil_secp256k1
// #cgo pkg-config: ${SRCDIR}/lib/pkgconfig/libfil_secp256k1.pc
// #include "./include/libfil_secp256k1.h"
import "C"

const PrivateKeyBytes = 32

// PublicKeySizeBytes is the size of a public key in ncompressed serialized format.
const PublicKeySizeBytes = 65

const MessageBytes = 32

type PrivateKey [PrivateKeyBytes]byte
type PublicKey struct{}

func PrivateKeyFromBytes(raw []byte) (*PrivateKey, error) {
	return nil, nil
}

func (pk *PrivateKey) PublicKey() *PublicKey {
	return nil
}

func (pk *PrivateKey) Sign(data []byte) ([]byte, error) {
	return nil, nil
}

func (pk *PrivateKey) Equals(other *PrivateKey) bool {
	return false
}

// Serialize serializes the key in uncompressed form.
func (pk *PublicKey) Serialize() []byte {
	return nil
}

func (pk *PublicKey) Verify(data, signature []byte) (bool, error) {
	return false, nil
}

func GenerateKeyFromSeed(seed io.Reader) (*PrivateKey, error) {
	return nil, nil
}

func GenerateKey() *PrivateKey {
	// call method
	resPtr := (*C.GenerateKeyResponse)(unsafe.Pointer(C.generate_key()))
	defer C.destroy_generate_key_response(resPtr)

	// prep response
	var key PrivateKey
	keySlice := C.GoBytes(unsafe.Pointer(&resPtr.key), PrivateKeyBytes)
	copy(key[:], keySlice)

	return &key
}

func EcRecover(data, signature []byte) (*PublicKey, error) {
	return nil, nil
}
