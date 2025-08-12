// Code generated from Pkl module `Diameter`. DO NOT EDIT.
package diameter

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Diameter struct {
	Apps []*App `pkl:"Apps"`

	Avps []*Avp `pkl:"Avps"`

	CmdFlags *CmdBitFlags `pkl:"CmdFlags"`

	AvpFlags *AvpBitFlags `pkl:"AvpFlags"`

	AvpTypes *AvpDataTypes `pkl:"AvpTypes"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Diameter
func LoadFromPath(ctx context.Context, path string) (ret *Diameter, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Diameter
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Diameter, error) {
	var ret Diameter
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
