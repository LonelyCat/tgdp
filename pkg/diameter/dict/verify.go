//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: verify.go
// Description: Diameter pkg: dictionary verification
//

package dict

import (
	"fmt"
)

func (d *Dict) Verify() int {
	errFound := 0

	for _, app := range d.core.GetApps() {
		for _, cmd := range app.Cmds {
			for _, rule := range cmd.Request {
				if avp, _ := d.GetAvpByName(rule.Name); avp == nil {
					fmt.Printf("%s/%sR: unknown AVP \"%s\"\n", app.Name, cmd.Short, rule.Name)
					errFound++
				}
			}
			for _, rule := range cmd.Answer {
				if avp, _ := d.GetAvpByName(rule.Name); avp == nil {
					fmt.Printf("%s/%sA: unknown AVP \"%s\"\n", app.Name, cmd.Short, rule.Name)
					errFound++
				}
			}
		}
	}
	fmt.Println()

	avpCodes := make(map[uint32]uint32)
	for _, avp := range d.core.GetAvps() {
		if code, exists := avpCodes[avp.Code]; exists {
			avpDup, _ := d.GetAvpByCode(code)
			fmt.Printf("Duplicated code %d for AVPs \"%s\" and \"%s\"\n", code, avp.Name, avpDup.Name)
			errFound++
		} else {
			avpCodes[avp.Code] = avp.Code
		}
	}
	fmt.Println()

	for _, avp := range d.core.GetAvps() {
		if avp.Flags&d.AvpFlag().V != 0 && avp.VndId == 0 {
			fmt.Printf("AVP \"%s\" V-flag persent without vendor id\n", avp.Name)
			errFound++
		}
	}
	fmt.Println()

	for _, avp := range d.core.GetAvps() {
		if avp.Type == d.AvpDataType().Grouped {
			for _, member := range avp.Group.Members {
				memberAvp, _ := d.GetAvpByName(member.Name)
				if memberAvp == nil {
					fmt.Printf("AVP \"%s\" unknown group member: \"%s\"\n", avp.Name, member.Name)
					errFound++
				}
			}
		}
	}
	fmt.Println()

	if errFound > 0 {
		fmt.Println(">>> Errors: ", errFound)
	} else {
		fmt.Println(">>> No errors foud :)")
	}
	return errFound
}
