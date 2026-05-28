//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: repl.go
// Description: REPL: start REPL mode
//

package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"tgdp/internal/config"
	"tgdp/internal/repl/avp"
	"tgdp/internal/repl/comp"
	"tgdp/internal/repl/echo"
	"tgdp/internal/repl/pcap"
	"tgdp/internal/repl/peer"
	"tgdp/internal/repl/receive"
	"tgdp/internal/repl/script"
	"tgdp/internal/repl/send"
	"tgdp/internal/repl/server"
	"tgdp/internal/repl/verbose"
	"tgdp/internal/repl/version"
	"tgdp/pkg/diameter"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// REPL commands
//

var (
	commands = []*cobra.Command{
		commandHelp,
		commandQuit,
		commandBatch,
		avp.RootCommand,
		echo.RootCommand,
		pcap.RootCommand,
		peer.RootCommand,
		receive.RootCommand,
		script.RootCommand,
		send.RootCommand,
		server.RootCommand,
		verbose.RootCommand,
		version.RootCommand,
	}

	commandBatch = &cobra.Command{
		Use:     "batch",
		Aliases: []string{"bat"},
		Short:   "batch <file> [file...]",
		Long:    "Execute commands from a batch file[s]",
		Example: "batch script.tgdp",
		Run:     batch,
	}

	commandHelp = &cobra.Command{
		Use:     "help",
		Aliases: []string{"?"},
		Short:   "help [command]",
		Long:    "Display help information",
		Example: "help send",
	}

	commandQuit = &cobra.Command{
		Use:     "quit",
		Aliases: []string{"exit", "bye"},
		Short:   "quit",
		Long:    "Quit from TGDP",
		Example: "quit",
		Run:     quit,
	}

	rootCommand = &cobra.Command{}
	forceQuit   = false
)

// Functions
//

func Run(env *diameter.Diameter) {
	for _, cmd := range commands {
		rootCommand.AddCommand(cmd)
		if cmd != commandHelp {
			commandHelp.AddCommand(cmd)
		}
	}
	commandHelp.Run = help

	bg := context.Background()
	ctx := context.WithValue(bg, diameter.EnvContext, env)
	rootCommand.SetContext(ctx)

	tgdpDir := config.DataDir()

	if _, err := os.Stat(tgdpDir); os.IsNotExist(err) {
		err = os.MkdirAll(tgdpDir, 0755)
		if err != nil {
			slog.Error("Error creating directory '%s': %s", tgdpDir, err)
		}
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:      "D> ",
		EOFPrompt:   commandQuit.Use,
		HistoryFile: config.HistoryFile(),
		AutoComplete: readline.NewPrefixCompleter(
			completionList(env)...,
		),
	})
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer func() {
		rl.Close() //nolint:errcheck
		env.Cancel()
		env.Wait()
		env.Peers().DisconnectAll(false)
	}()

	batch(nil, []string{config.AutoRunFile()})

	for !forceQuit {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		}

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		args := strings.Fields(input)
		rootCommand.ResetFlags()
		rootCommand.SetArgs(args)
		if err := rootCommand.Execute(); err != nil {
			fmt.Println(err)
		}

		rl.Refresh()
	}
}

func completionList(env *diameter.Diameter) []readline.PrefixCompleterInterface {
	pciList := []readline.PrefixCompleterInterface{}

	pciList = append(pciList, readline.PcItem(commandHelp.Use, comp.SubList(rootCommand)...))
	pciList = append(pciList, readline.PcItem(commandQuit.Use))
	pciList = append(pciList, readline.PcItem(commandBatch.Use, comp.FileList(config.BatchDir())...))
	pciList = append(pciList, avp.CompList(env)...)
	pciList = append(pciList, echo.CompList()...)
	pciList = append(pciList, pcap.CompList(env)...)
	pciList = append(pciList, peer.CompList(env)...)
	pciList = append(pciList, receive.CompList(env)...)
	pciList = append(pciList, script.CompList()...)
	pciList = append(pciList, send.CompList(env)...)
	pciList = append(pciList, server.CompList()...)
	pciList = append(pciList, verbose.CompList()...)
	pciList = append(pciList, version.CompList()...)

	return pciList
}

func quit(cmd *cobra.Command, args []string) {
	forceQuit = true
}

func help(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Available commands:")
		for _, cmd := range commands {
			fmt.Printf("  %-8s - %s\n", cmd.Use, cmd.Long)
		}
		return
	}
	helpCommand(cmd, args)
}

func helpCommand(cmd *cobra.Command, args []string) {
	for _, sub := range cmd.Commands() {
		for i, arg := range args {
			if strings.EqualFold(arg, sub.Use) {
				if i < len(args)-1 {
					helpCommand(sub, args[i:])
				} else {
					sub.Help() //nolint:errcheck
				}
				return
			}
		}
		fmt.Printf("Unknown command '%s'\n", args[0])
	}
}

func batch(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println(cmd.Short)
		return
	}

	for _, file := range args {
		if file[0] != '/' && file[0] != '.' {
			file = filepath.Join(config.BatchDir(), file)
		}

		fd, err := os.Open(file)
		if err != nil {
			fmt.Println(err)
			continue
		}

		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			line := scanner.Text()
			input := strings.TrimSpace(line)
			if input == "" || input[0] == '#' {
				continue
			}

			args := strings.Fields(input)
			rootCommand.SetArgs(args)
			if err := rootCommand.Execute(); err != nil {
				fmt.Println(err)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}

		fd.Close() //nolint:errcheck
	}
}
