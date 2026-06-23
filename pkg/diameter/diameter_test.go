package diameter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"tgdp/pkg/diameter/dict"
)

var dia *Diameter

func TestDict(t *testing.T) {
	fmt.Println(">>> Dictionary test")
	d, err := New(ModeTransaction)
	if err != nil {
		t.Fatal(err)
		return
	}
	dia = d

	err = dia.LoadDict("./dict/pkl/dictionary.pkl", dict.FormatPkl)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print("Dictionary loaded succsessfuly\n\n")
	dia.Dict().Show()

	fmt.Println("<<< Dictionary test")
}

func TestDataStore(t *testing.T) {
	fmt.Println(">>> Data Store test")

	avpData := filepath.Join(os.Getenv("HOME"), ".tgdp/data/avp-data.yaml")
	err := dia.LoadData(avpData)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print("AVP data loaded succsessfuly\n\n")
	dia.Store().Dump(0)

	fmt.Println("<<< Data Store test")
}

func TestMessage(t *testing.T) {
	fmt.Println(">>> Message test")

	msg, err := dia.NewMessage("S6a", "UL", true, true)
	if err != nil {
		t.Fatal(err)
		return
	}
	msg.Trace(0)

	avp, err := msg.GetAvp("User-Name")
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Printf("User-Name: %v\n", avp.Value())

	fmt.Println()

	msg, err = dia.NewMessage("S6a", "CL", false, true)
	if err != nil {
		t.Fatal(err)
		return
	}

	avp, err = msg.GetAvp("Result-Code")
	if err != nil {
		t.Fatal(err)
		return
	}
	var rc any = uint32(9999)
	err = avp.SetValue(rc)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println()
	msg.Trace(0)

	fmt.Println("<<< Message test")
}

func TestPeers(t *testing.T) {
	fmt.Println(">>> Peers test")

	peersData := filepath.Join(os.Getenv("HOME"), ".tgdp/data/peers.yaml")
	err := dia.LoadPeers(peersData)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print("Peers loaded succsessfuly\n\n")
	dia.Peers().Dump(0)

	fmt.Println("<<< Peers test")
}

func TestNode(t *testing.T) {
	fmt.Println(">>> Node test")

	peer, err := dia.NewPeer("TestPeer", "localhost", 3868, "tcp", 0)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println("Peer created successfully")

	dia.SetTraceLevel(TraceCM)

	err = peer.Connect()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println("Peer connected successfully")
	peer.Trace()

	msg, err := dia.NewRequest("S6a", "UL")
	if err != nil {
		t.Fatal(err)
		return
	}
	msg.Trace()

	_, err = msg.Serialize()
	if err != nil {
		t.Fatal(err)
		return
	}

	err = peer.SendTo(msg.Bytes())
	if err != nil {
		t.Fatal(err)
		return
	}

	msg, err = dia.RecvMessage(peer, true)
	if err != nil {
		t.Fatal(err)
		return
	}
	msg.Trace()

	err = peer.Disconnect()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println("Peer disconnected successfully")

	fmt.Println(">>> Node test")
}
