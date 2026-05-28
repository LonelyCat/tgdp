// Code generated from Pkl module `Core`. DO NOT EDIT.
package dict

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Core interface {
	Diameter

	GetApps() []App

	GetAvps() []Avp

	GetCmdFlags() CmdBitFlags

	GetAvpFlags() AvpBitFlags

	GetAvpTypes() AvpDataTypes
}

var _ Core = CoreImpl{}

type CoreImpl struct {
	Apps []App `pkl:"apps"`

	Avps []Avp `pkl:"avps"`

	CmdFlags CmdBitFlags `pkl:"cmdFlags"`

	AvpFlags AvpBitFlags `pkl:"avpFlags"`

	AvpTypes AvpDataTypes `pkl:"avpTypes"`
}

func (rcv CoreImpl) GetApps() []App {
	return rcv.Apps
}

func (rcv CoreImpl) GetAvps() []Avp {
	return rcv.Avps
}

func (rcv CoreImpl) GetCmdFlags() CmdBitFlags {
	return rcv.CmdFlags
}

func (rcv CoreImpl) GetAvpFlags() AvpBitFlags {
	return rcv.AvpFlags
}

func (rcv CoreImpl) GetAvpTypes() AvpDataTypes {
	return rcv.AvpTypes
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Core
func LoadFromPath(ctx context.Context, path string) (ret Core, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ret, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Core
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Core, error) {
	var ret CoreImpl
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
