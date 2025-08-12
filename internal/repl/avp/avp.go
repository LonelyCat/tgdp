//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: avp.go
// Description: REPL: 'avp' command implementation
//

package avp

import (
	"fmt"
	"path/filepath"
	"strconv"

	"tgdp/internal/config"
	"tgdp/internal/repl/comp"
	"tgdp/pkg/diameter"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// -- Constants
// --
const padOffset = 2

// -- Variables
// --
var (
	RootCommand = &cobra.Command{
		Use:   "avp",
		Short: "avp <info | get | set | add | delete> <id | name> [<index> <value>]",
		Long:  "Manage AVPs & global values",
	}

	SubCommandInfo = &cobra.Command{
		Use:     "info",
		Short:   "avp info <id | name>",
		Long:    "Show info about the AVP",
		Example: "avp info User-Name",
		Run:     info,
	}

	SubCommandGet = &cobra.Command{
		Use:     "get",
		Short:   "avp get <id | name>",
		Long:    "Get AVP value",
		Example: "avp get 264",
		Run:     get,
	}

	SubCommandSet = &cobra.Command{
		Use:     "set",
		Short:   "avp set <id | name> <index> <value>",
		Long:    "Set AVP value",
		Example: "avp set Origin-Host 0 dra01.epc.mnc000.mcc000.3gppnetwork.org",
		Run:     set,
	}

	SubCommandAdd = &cobra.Command{
		Use:     "add",
		Short:   "avp add <id | name> <value>",
		Long:    "Add value to AVP",
		Example: "avp add Auth-Application-Id 16777218",
		Run:     add,
	}

	SubCommandDel = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del", "rm"},
		Short:   "avp delete <id | name> <index>",
		Long:    "Delete value to AVP",
		Example: "avp del 258 2",
		Run:     del,
	}

	SubCommandLoad = &cobra.Command{
		Use:     "load",
		Short:   "avp load [flags] <file.yaml>",
		Long:    "Load AVP data from a YAML file",
		Example: "avp load subs-data.yaml",
		Run:     load,
	}
)

// -- Functions
// --
func init() {
	RootCommand.AddCommand(SubCommandInfo)
	RootCommand.AddCommand(SubCommandGet)
	RootCommand.AddCommand(SubCommandSet)
	RootCommand.AddCommand(SubCommandAdd)
	RootCommand.AddCommand(SubCommandDel)
	RootCommand.AddCommand(SubCommandLoad)
}

// -- Functions
// --
func CompList() []readline.PrefixCompleterInterface {
	pciSub := []readline.PrefixCompleterInterface{}

	for _, sub := range RootCommand.Commands() {
		switch sub.Use {
		case "load":
			pciSub = append(pciSub, readline.PcItem(sub.Use, comp.FileList(config.YamlDir())...))
		default:
			pciSub = append(pciSub, readline.PcItem(sub.Use, comp.AvpList()...))
		}
	}

	return []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciSub...)}
}

func info(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(cmd.Short)
		return
	}

	avp, err := diameter.Dict.GetAvp(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Code: %d\n", avp.Code)
	fmt.Printf("Name: %s\n", avp.Name)
	if avp.VndId != 0 {
		fmt.Printf("Vendor ID: %d\n", avp.VndId)
	}
	fmt.Printf("Type: %s\n", diameter.Dict.AvpTypes.String(avp.Type))
	switch avp.Type {
	case diameter.Dict.AvpTypeEnumerated():
		showEnumeratedInfo(avp)
	case diameter.Dict.AvpTypeGrouped():
		showGroupedInfo(avp, 0)
	}
}

func get(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(cmd.Short)
		return
	}

	getAvpValues(args[0], nil, 0)
}

func set(cmd *cobra.Command, args []string) {
	if len(args) < 3 {
		fmt.Println(cmd.Short)
		return
	}

	setAvpValue(args)
	getAvpValues(args[0], nil, 0)
}

func add(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println(cmd.Short)
		return
	}

	yamlText := fmt.Sprintf("%s: %s", args[0], args[1])
	if err := diameter.AvpDataFromYamlStr(yamlText, -1); err != nil {
		fmt.Println(err)
	}

	getAvpValues(args[0], nil, 0)
}

func del(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println(cmd.Short)
		return
	}

	index := diameter.AVP_DATA_CLEANUP
	if len(args) > 1 {
		i, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(cmd.Short)
			return
		}
		index = i
	}

	if err := diameter.DelAvpValue(args[0], index, nil); err != nil {
		fmt.Println(err)
	}

	getAvpValues(args[0], nil, 0)
}

func load(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println(cmd.Short)
	}

	for _, file := range args {
		if file[0] != '/' && file[0] != '.' {
			file = filepath.Join(config.YamlDir(), file)
		}
		err := diameter.LoadAvpsDataFromYaml(file, diameter.AVP_DATA_APPEND)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func showGroupedInfo(avp *diameter.Avp, shift int) {
	fmt.Println("Members:")
	for _, member := range avp.Group.Members {
		for range shift {
			fmt.Print(" ")
		}

		if member.Required {
			fmt.Print("* ")
		} else {
			fmt.Print("  ")
		}

		fmt.Printf(" %2d %s\n", member.Max, member.Name)
	}
}

func showEnumeratedInfo(avp *diameter.Avp) {
	fmt.Println("Values:")
	for _, item := range avp.Enum.Items {
		fmt.Printf("  %s (%d)\n", item.Name, item.Code)
	}
}

func getAvpValues(avpId any, store *diameter.AvpDataStore, shift int) {
	avp, data, err := diameter.FetchAvpValue2(avpId, store)
	if err != nil {
		fmt.Println(err)
		return
	}

	pad := func(s int) {
		for range s {
			fmt.Print(" ")
		}
	}

	pad(shift)
	fmt.Printf("%s (%d):\n", avp.Name, avp.Code)
	for i, value := range data {
		switch avp.Type {
		case diameter.Dict.AvpTypeEnumerated():
			pad(shift)
			getEnumeratedValue(i, avp, value.Value)
		case diameter.Dict.AvpTypeGrouped():
			getGroupedValue(value, shift+padOffset)
		default:
			pad(shift)
			if value != nil {
				fmt.Printf("%3d: %v\n", i, value.Value)
			} else {
				fmt.Printf("%3d: <undefined>\n", i)
			}
		}
	}
}

func getEnumeratedValue(index int, avp *diameter.Avp, value any) {
	for _, item := range avp.Enum.Items {
		if item.Code == value.(int32) {
			fmt.Printf("%3d: %s (%d)\n", index, item.Name, item.Code)
			return
		}
	}
	fmt.Printf("%3d, %s(%d) = %v\n", index, avp.Name, avp.Code, value)
}

func getGroupedValue(value *diameter.AvpData, shift int) {
	store := value.Value.(*diameter.AvpDataStore)
	for id := range *store {
		getAvpValues(id, store, shift)
	}
}

func setAvpValue(args []string) {
	index, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	yamlData := fmt.Sprintf("%s: %s", args[0], args[2])
	if err := diameter.AvpDataFromYamlStr(yamlData, index); err != nil {
		fmt.Println(err)
	}
}
