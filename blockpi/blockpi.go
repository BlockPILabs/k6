package blockpi

import "go.k6.io/k6/js/modules"
import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct{}

	// BlockPI represents an instance of the k6 module.
	BlockPI struct {
		vu modules.VU
	}
)

// New returns a pointer to a new RootModule instance.
func New() *RootModule {
	return &RootModule{}
}

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &BlockPI{vu: vu}
}

// Exports returns the exports of the k6 module.
func (pi *BlockPI) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"sign": pi.sign,
		},
	}
}

// sha1 returns the SHA1 hash of input in the given encoding.
func (pi *BlockPI) sign(privateKeyHex string, data string) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err

	}
	hash := crypto.Keccak256Hash([]byte(data))
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signature), nil

}
