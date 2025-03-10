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
	tempfile, error := getTempFile()
	if error != nil {
		fmt.Fprintf(os.Stderr, "%s\n", error.Error())
		return 1
	}

	// Make sure we clean up the file
	defer func() {
		error := CleanupFiles(tempfile)
		if error != nil {
			fmt.Fprintf(os.Stderr, "Failed to cleanup at %s\n%s\n", tempfile.DirName, error.Error())
		}
	}()

	fmt.Printf("Created temp file at %s\n", tempfile.File.Name())

	err := SetupPlaygroundConfig(tempfile.DirName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return 1
	}

	fmt.Printf("Setup playground at %s\n", tempfile.DirName)

	// Run TS Playground
	{
		err := OpenNeovim(tempfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return 1
		}

		err = CommandLoop(tempfile)
		if err != nil {
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

func CommandLoop(playgroundFiles *PlaygroundFiles) error {
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
				out := RunScript(playgroundFiles)
				fmt.Printf("\n%s", out)
				break
			}
		case CommandEditor:
			{
				if err := OpenNeovim(playgroundFiles); err != nil {
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

func OpenNeovim(playgroundFiles *PlaygroundFiles) error {
	relativeFilename := playgroundFiles.RelativeFilename()

	cmd := exec.Command("nvim", relativeFilename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = playgroundFiles.DirName

	err := cmd.Run()

	if err != nil {
		return handleExecError(
			err,
			fmt.Sprintf("Failed to open neovim in %s", playgroundFiles.File.Name()),
		)
	}

	return nil
}

func RunScript(playgroundFiles *PlaygroundFiles) string {
	return "TODO: implement running script\n\n"
}

func SetupPlaygroundConfig(dirname string) error {
	// create tsconfig
	{
		cmd := exec.Command("npx", "-y", "tsc", "--init")
		cmd.Dir = dirname

		if out, err := cmd.Output(); err != nil {
			fmt.Println(out)
			return handleExecError(err, "Failed to setup playground project")
		}
	}

	return nil
}

type PlaygroundFiles struct {
	File    *os.File
	DirName string
}

func (pf *PlaygroundFiles) RelativeFilename() string {
	if pf.File == nil {
		return ""
	}

	_, filename := path.Split(pf.File.Name())
	return filename
}

func CleanupFiles(playgroundFiles *PlaygroundFiles) error {
	var err error
	if playgroundFiles.File != nil {
		_ = playgroundFiles.File.Close()
	}

	if playgroundFiles.DirName != "" {
		err = os.RemoveAll(playgroundFiles.DirName)
	}

	return err
}

func getTempFile() (*PlaygroundFiles, error) {
	osTempDir := os.TempDir()
	if osTempDir == "" {
		return nil, errors.New("failed to retrieve temporary directory")
	}

	out := PlaygroundFiles{}

	var err error
	out.DirName, err = os.MkdirTemp(osTempDir, "ts-playground-*")
	if err != nil {
		return nil, errors.Join(
			errors.New("failed to create temporary directory"),
			err,
		)
	}

	tempfileName := path.Join(out.DirName, "playground.ts")
	out.File, err = os.Create(tempfileName)
	if err != nil {
		CleanupFiles(&out)
		return nil, errors.Join(
			errors.New("failed to create temporary file"),
			err,
		)
	}

	err = os.Chmod(out.File.Name(), os.ModePerm)
	if err != nil {
		CleanupFiles(&out)
		return nil, errors.Join(
			errors.New("failed to get allocated permission on playground file"),
			err,
		)
	}

	return &out, nil
}

// log stderr and wrap the error messaging
func handleExecError(err error, message string) error {
	if exitErr, isExitErr := err.(*exec.ExitError); isExitErr {
		fmt.Fprint(os.Stderr, string(exitErr.Stderr))
	}

	return errors.Join(
		errors.New(message),
		err,
	)
}
