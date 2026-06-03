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
	"slices"
	"sort"
	"strconv"

	"tgdp/internal/config"
	"tgdp/internal/repl/comp"
	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/dict"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// Constants
//

const padOffset = 2

// Variables
//

var (
	RootCommand = &cobra.Command{
		Use:   "avp",
		Short: "avp <list | info | get | set | add | delete | load | purge> [parameters]",
		Long:  "Manage AVP global values",
	}

	SubCommandList = &cobra.Command{
		Use:     "list",
		Short:   "avp list",
		Long:    "Show list of AVP values",
		Example: "avp list",
		Run:     list,
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

	SubCommandPurge = &cobra.Command{
		Use:     "purge",
		Short:   "avp purge",
		Long:    "Purge all AVP",
		Example: "avp purge",
		Run:     purge,
	}
)

// Functions
//

func CompList(env *diameter.Diameter) []readline.PrefixCompleterInterface {
	pciSub := []readline.PrefixCompleterInterface{}

	for _, sub := range RootCommand.Commands() {
		switch sub {
		case SubCommandList:
			pciSub = append(pciSub, readline.PcItem(sub.Use))
		case SubCommandLoad:
			pciSub = append(pciSub, readline.PcItem(sub.Use, comp.FileList(config.YamlDir())...))
		default:
			pciSub = append(pciSub, readline.PcItem(sub.Use, comp.AvpList(env)...))
		}
	}

	return []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciSub...)}
}

func list(cmd *cobra.Command, args []string) {
	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)
	listAvpsData(env.Store(), 0)
}

func info(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	avp, err := env.Dict().GetAvp(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Code: %d\n", avp.Code)
	fmt.Printf("Name: %s\n", avp.Name)
	if avp.VndId != 0 {
		fmt.Printf("Vendor ID: %d\n", avp.VndId)
	}
	fmt.Printf("Type: %s\n", env.Dict().AvpDataTypeName(avp.Type))
	switch avp.Type {
	case env.Dict().AvpDataType().Enumerated:
		showEnumeratedInfo(avp)
	case env.Dict().AvpDataType().Grouped:
		showGroupedInfo(avp, 0)
	}
}

func get(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	getAvpValues(env, args[0], 0)
}

func set(cmd *cobra.Command, args []string) {
	if len(args) < 3 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	setAvpValue(env, args)
	getAvpValues(env, args[0], 0)
}

func add(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	yamlText := fmt.Sprintf("%s: %s", args[0], args[1])
	if err := env.Store().MakeFromYaml(yamlText, diameter.AvpStoreAppend, 0); err != nil {
		fmt.Println(err)
	}

	getAvpValues(env, args[0], 0)
}

func del(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	avpCode, err := strconv.Atoi(args[0])
	if err != nil {

		if avp, err := env.Dict().GetAvp(args[0]); err != nil {
			fmt.Println(err)
			return
		} else {
			avpCode = int(avp.Code)
		}

	}

	index, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println(cmd.Short)
		return
	}

	if ok := env.Store().Delete(uint32(avpCode), index); !ok {
		fmt.Printf("Delete failed: avp %d not found or wrong index %d\n", avpCode, index)
	}

	getAvpValues(env, args[0], 0)
}

func load(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	for _, file := range args {
		if file[0] != '/' && file[0] != '.' {
			file = filepath.Join(config.YamlDir(), file)
		}
		err := env.Store().LoadFromFile(file, diameter.AvpStoreAppend, 0)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func purge(cmd *cobra.Command, args []string) {
	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)
	env.Store().Purge()
}

// Helpers
//

func listAvpsData(store *diameter.AvpStore, shift int) {
	var list []*diameter.Avp
	for _, avps := range slices.Collect(store.Iter()) {
		list = append(list, avps...)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})

	for _, avp := range list {
		showAvpData(avp, shift)
	}
}

func showAvpData(avp *diameter.Avp, shift int) {
	for range shift {
		fmt.Print(" ")
	}

	fmt.Printf("%s (%d)", avp.Name(), avp.Code())
	if avp.IsGrouped() {
		fmt.Println()
		for _, member := range avp.Value().([]*diameter.Avp) {
			showAvpData(member, shift+2)
		}
	} else {
		if codec, exists := avp.Env().Codec(avp.Type()); exists {
			fmt.Printf(" = %s\n", codec.ToText(avp))
		} else {
			fmt.Printf(" = %v\n", avp.Value())
		}
	}
}

func showGroupedInfo(avp *dict.Avp, shift int) {
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

		if member.Max != nil {
			fmt.Printf("  %2d %s\n", *member.Max, member.Name)
		} else {
			fmt.Printf("  -- %s\n", member.Name)
		}
	}
}

func showEnumeratedInfo(avp *dict.Avp) {
	fmt.Println("Values:")
	for _, item := range avp.Enum.Items {
		fmt.Printf("  %s (%d)\n", item.Name, item.Code)
	}
}

func getAvpValues(env *diameter.Diameter, avpId any, shift int) {
	avp, err := env.Dict().GetAvp(avpId)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := env.Store().Fetch(avp.Code)
	if data == nil {
		return
	}

	pad := func(s int) {
		for range s {
			fmt.Print(" ")
		}
	}

	pad(shift)
	fmt.Printf("%s (%d):\n", avp.Name, avp.Code)
	for _, value := range data {
		showAvpData(value, shift+padOffset)
	}
}

func setAvpValue(env *diameter.Diameter, args []string) {
	index, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	yamlData := fmt.Sprintf("%s: %s", args[0], args[2])
	if err := env.Store().MakeFromYaml(yamlData, diameter.AvpStoreReplace, index); err != nil {
		fmt.Println(err)
	}
}

// Init
//

func init() {
	RootCommand.AddCommand(SubCommandList)
	RootCommand.AddCommand(SubCommandInfo)
	RootCommand.AddCommand(SubCommandGet)
	RootCommand.AddCommand(SubCommandSet)
	RootCommand.AddCommand(SubCommandAdd)
	RootCommand.AddCommand(SubCommandDel)
	RootCommand.AddCommand(SubCommandLoad)
	RootCommand.AddCommand(SubCommandPurge)
}
