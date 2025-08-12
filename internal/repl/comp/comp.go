//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: comp.go
// Description: REPL: completions for dynamic objects
//

package comp

import (
	"os"
	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/net/node"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

func SubList(cmd *cobra.Command) []readline.PrefixCompleterInterface {
	pciList := []readline.PrefixCompleterInterface{}

	for _, sub := range cmd.Commands() {
		pciList = append(pciList, SubList(sub)...)
	}

	if cmd.Use == "" {
		return pciList
	}

	return []readline.PrefixCompleterInterface{readline.PcItem(cmd.Use, pciList...)}
}

func FileList(dir string) []readline.PrefixCompleterInterface {
	filesNames := func() func(string) []string {
		return func(line string) []string {
			files, err := os.ReadDir(dir)
			if err != nil {
				return nil
			}

			names := []string{}
			for _, file := range files {
				if file.Name()[0] == '.' {
					continue
				}
				if file.IsDir() {
					// TODO: Implement directory completion
					// pciList = []string{}
				} else {
					names = append(names, file.Name())
				}
			}
			return names
		}
	}

	return []readline.PrefixCompleterInterface{readline.PcItemDynamic(filesNames())}
}

func PeerList(incApps bool) []readline.PrefixCompleterInterface {
	peerNames := func() func(string) []string {
		return func(line string) []string {
			names := []string{}
			for peer := range node.Iter() {
				names = append(names, peer.Name)
			}
			return names
		}
	}

	if incApps {
		return []readline.PrefixCompleterInterface{readline.PcItemDynamic(peerNames(), AppList(true)...)}
	}
	return []readline.PrefixCompleterInterface{readline.PcItemDynamic(peerNames())}
}

func AppList(incCmds bool) []readline.PrefixCompleterInterface {
	pciList := []readline.PrefixCompleterInterface{}

	for app := range diameter.Dict.AppIter() {
		if incCmds {
			pciList = append(pciList, readline.PcItem(app.Name, CmdList(app)...))
		} else {
			pciList = append(pciList, readline.PcItem(app.Name))
		}
	}

	return pciList
}

func CmdList(app *diameter.App) []readline.PrefixCompleterInterface {
	pciList := []readline.PrefixCompleterInterface{}

	for cmd := range app.CmdIter() {
		pciList = append(pciList, readline.PcItem(cmd.Short))
	}

	return pciList
}

func AvpList() []readline.PrefixCompleterInterface {
	avpsNames := func() func(string) []string {
		return func(line string) []string {
			names := []string{}
			for avp := range diameter.Dict.AvpIter() {
				names = append(names, avp.Name)
			}
			return names
		}
	}

	return []readline.PrefixCompleterInterface{readline.PcItemDynamic(avpsNames())}
}

func AvpDataList() []readline.PrefixCompleterInterface {
	avpsNames := func() func(string) []string {
		return func(line string) []string {
			names := []string{}
			for code, _ := range diameter.AvpDataIter2(nil) {
				avp, err := diameter.Dict.GetAvpByCode(code)
				if err != nil {
					continue
				}
				names = append(names, avp.Name)
			}
			return names
		}
	}

	return []readline.PrefixCompleterInterface{readline.PcItemDynamic(avpsNames())}
}
