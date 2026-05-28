package server

import (
	"os"
	"os/signal"
	"testing"

	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/dict"
)

func TestServer(t *testing.T) {
	d, err := diameter.New(diameter.ModeTransaction)
	if err != nil {
		t.Fatal(err)
	}

	err = d.LoadDict("../../dict/pkl/dictionary.pkl", dict.FormatPkl)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = d.LoadData("../../test/avp-data.yaml")
	if err != nil {
		t.Fatal(err)
		return
	}

	d.SetTraceLevel(diameter.TraceCM)

	s := New(d)

	ccChan := make(chan os.Signal, 1)
	signal.Notify(ccChan, os.Interrupt)
	go func() {
		<-ccChan
		s.Shutdown()
	}()

	s.SetAutoreply(true)
	s.SetVerboseLevel(Info)
	if err := s.Start("localhost:3868", true); err != nil {
		t.Fatal(err)
		return
	}

	s.Dump()
	s.Wait()
	s.Dump()
}
