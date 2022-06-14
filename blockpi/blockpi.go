package blockpi

import (
	"errors"
	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/metrics"
	"sync"
	"sync/atomic"
)
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
	xatomic struct {
		v *int64
	}
)

var atomics = &sync.Map{}

func newAtomic() *xatomic {
	a := &xatomic{v: new(int64)}
	*a.v = 0
	return a
}

func (a *xatomic) Add(v goja.Value) (int64, error) {
	vfloat := v.ToInteger()
	if vfloat == 0 && v.ToBoolean() {
		vfloat = 1.0
	}
	return atomic.AddInt64(a.v, vfloat), nil
}

func (a *xatomic) Get() (int64, error) {
	return *a.v, nil
}

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
			"sign":   pi.sign,
			"Atomic": pi.XAtomic,
		},
	}
}

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

// sha1 returns the SHA1 hash of input in the given encoding.
func (pi *BlockPI) XAtomic(call goja.ConstructorCall, rt *goja.Runtime) *goja.Object {

	v, err := pi.newXAtomic(call, metrics.Counter)
	if err != nil {
		common.Throw(rt, err)
	}
	return v
}

// sha1 returns the SHA1 hash of input in the given encoding.
func (mi *BlockPI) newXAtomic(call goja.ConstructorCall, t metrics.MetricType) (*goja.Object, error) {
	initEnv := mi.vu.InitEnv()
	if initEnv == nil {
		return nil, errors.New("metrics must be declared in the init context")
	}
	rt := mi.vu.Runtime()
	c, _ := goja.AssertFunction(rt.ToValue(func(name string, isTime ...bool) (*goja.Object, error) {
		a, _ := atomics.LoadOrStore(name, newAtomic())
		m, _ := a.(*xatomic)

		o := rt.NewObject()
		err := o.DefineDataProperty("name", rt.ToValue(name), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_TRUE)
		if err != nil {
			return nil, err
		}
		if err = o.Set("add", rt.ToValue(m.Add)); err != nil {
			return nil, err
		}
		if err = o.Set("get", rt.ToValue(m.Get)); err != nil {
			return nil, err
		}
		return o, nil
	}))
	v, err := c(call.This, call.Arguments...)
	if err != nil {
		return nil, err
	}

	return v.ToObject(rt), nil
}
