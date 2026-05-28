package repl

import (
	"fmt"
	"testing"

	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/dict"
)

func TestRepl(t *testing.T) {
	d, err := diameter.New(diameter.ModeTransaction)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = d.LoadDict("../../pkg/diameter/dict/pkl/dictionary.pkl", dict.FormatPkl)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print("Dictionary loaded succsessfuly\n\n")

	err = d.LoadData("../../pkg/diameter/test/avp-data.yaml")
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print("AVP data loaded succsessfuly\n\n")

	err = d.LoadPeers("../../pkg/diameter/test/peers.yaml")
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print("Peers loaded succsessfuly\n\n")

	Run(d)
}
