package dict

import (
	"fmt"
	"testing"
)

var dic Dict

func TestLoadFromFile(t *testing.T) {
	err := dic.LoadFromFile("./pkl/dictionary.pkl", FormatPkl)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Print("Dictionary loaded succsessfuly\n\n")
}

func TestCmdFlags(t *testing.T) {
	fmt.Println("Command flags:")
	fmt.Printf("%3d - %s\n", dic.CmdFlag().R, dic.CmdFlagName(dic.CmdFlag().R))
	fmt.Printf("%3d - %s\n", dic.CmdFlag().P, dic.CmdFlagName(dic.CmdFlag().P))
	fmt.Printf("%3d - %s\n", dic.CmdFlag().E, dic.CmdFlagName(dic.CmdFlag().E))
	fmt.Printf("%3d - %s\n", dic.CmdFlag().T, dic.CmdFlagName(dic.CmdFlag().T))
	fmt.Println()
}

func TestAvpFlags(t *testing.T) {
	fmt.Println("AVP flags:")
	fmt.Printf("%3d - %s\n", dic.AvpFlag().V, dic.AvpFlagName(dic.AvpFlag().V))
	fmt.Printf("%3d - %s\n", dic.AvpFlag().M, dic.AvpFlagName(dic.AvpFlag().M))
	fmt.Printf("%3d - %s\n", dic.AvpFlag().P, dic.AvpFlagName(dic.AvpFlag().P))
	fmt.Println()
}

func TestAvpTypes(t *testing.T) {
	fmt.Println("AVP types:")
	for i := 1; i <= 16; i++ {
		fmt.Printf("%2d - %s\n", i, dic.AvpDataTypeName(i))
	}
	fmt.Println()
}

func TestAvp(t *testing.T) {
	avp, err := dic.GetAvp(1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("AVP: %s(%d)\n", avp.Name, avp.Code)

	avp, err = dic.GetAvp("MSISDN")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("AVP: %s(%d)\n", avp.Name, avp.Code)

	_, err = dic.GetAvp("Not-Existing")
	if err != nil {
		fmt.Println(err)
	}
}

// func TestShow(t *testing.T) {
// 	dic.Show()
// }

// func TestDump(t *testing.T) {
// 	dic.Dump()
// }

// func TestVerify(t *testing.T) {
// 	dic.Verify()
// }
