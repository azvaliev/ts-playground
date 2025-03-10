package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	status := Run()
	os.Exit(status)
}

func Run() int {
	playground, error := setupTempFiles()

	// Make sure we clean up the playground
	defer func() {
		if err := playground.Destroy(); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Failed to cleanup at %s\n%s\n",
				playground.DirName,
				error.Error(),
			)
		}
	}()

	if error != nil {
		fmt.Fprintf(os.Stderr, "%s\n", error.Error())
		return 1
	}

	// Run TS Playground
	{
		if err := OpenNeovim(playground); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return 1
		}

		if err := CommandLoop(playground); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return 1
		}
	}

	return 0
}

type Command rune

const (
	CommandRun    Command = 'r'
	CommandEditor Command = 'e'
)

func CommandLoop(playground *Playground) error {
	for {
		fmt.Printf(
			"Enter command (%s to run, %s for editor, ctrl-c to exit)\n"+
				"> ",
			string(CommandRun),
			string(CommandEditor),
		)

		var cmd string
		_, err := fmt.Scan(&cmd)

		if err != nil || cmd == "" {
			return errors.Join(
				errors.New("could not read input"),
				err,
			)
		}

		switch Command(cmd[0]) {
		case CommandRun:
			{
				out := RunScript(playground)
				fmt.Printf("\n%s", out)
				break
			}
		case CommandEditor:
			{
				if err := OpenNeovim(playground); err != nil {
					return err
				}
				break
			}
		default:
			{
				fmt.Printf("Unknown command %s\n", cmd)
			}
		}
	}
}

func OpenNeovim(playground *Playground) error {
	relativeFilename := playground.GetRelativeFilename()

	cmd := exec.Command("nvim", relativeFilename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = playground.DirName

	err := cmd.Run()

	if err != nil {
		return handleExecError(
			err,
			fmt.Sprintf("Failed to open neovim in %s", playground.File.Name()),
		)
	}

	return nil
}

func RunScript(playground *Playground) string {
	return "TODO: implement running script\n\n"
}
