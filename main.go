package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
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

type Command string

const (
	CommandRun    Command = "r"
	CommandEditor Command = "e"
	CommandClear  Command = "clear"
)

func CommandLoop(playground *Playground) error {
	for {
		fmt.Printf(
			"Enter command (%s to run, %s for editor, %s to clear output, ctrl-c to exit)\n"+
				"> ",
			string(CommandRun),
			string(CommandEditor),
			string(CommandClear),
		)

		var cmd string
		_, err := fmt.Scan(&cmd)

		if err != nil || cmd == "" {
			return errors.Join(
				errors.New("could not read input"),
				err,
			)
		}

		switch Command(cmd) {
		case CommandRun:
			{
				RunScript(playground)
				break
			}
		case CommandEditor:
			{
				if err := OpenNeovim(playground); err != nil {
					return err
				}
				break
			}
		case CommandClear:
			{
				ClearStdout()
				break
			}
		default:
			{
				fmt.Printf("Unknown command %s\n", cmd)
				break
			}
		}

		fmt.Print("\n")
	}
}

func OpenNeovim(playground *Playground) error {
	relativeFilename := playground.GetRelativeFilename()

	cmd := exec.Command("nvim", relativeFilename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = playground.DirName

	if err := cmd.Run(); err != nil {
		return handleExecError(
			err,
			fmt.Sprintf("Failed to open neovim in %s", playground.File.Name()),
		)
	}

	return nil
}

func RunScript(playground *Playground) error {
	pathToInputFile := playground.File.Name()
	pathToOutFile := path.Join(playground.DirName, "playground.js")

	// Compile the TypeScript file
	{
		tsc := "tsc"
		cmd := exec.Command(
			"npx",
			tsc,
			pathToInputFile,
			"--outFile",
			pathToOutFile,
		)
		cmd.Dir = playground.DirName
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if exitError, isExitError := err.(*exec.ExitError); isExitError {
			return fmt.Errorf("%s exited with status code %d", tsc, exitError.ExitCode())
		} else if err != nil {
			return err
		}
	}

	// Execute the compiled JavaScript file
	{
		node := "node"
		cmd := exec.Command(node, pathToInputFile)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = playground.DirName

		err := cmd.Run()
		if exitError, isExitError := err.(*exec.ExitError); isExitError {
			return fmt.Errorf("%s exited with status code %d", node, exitError.ExitCode())
		} else if err != nil {
			return err
		}
	}

	return nil
}

func ClearStdout() {
	fmt.Print("\033[2J")
}
