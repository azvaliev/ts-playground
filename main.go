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

	return 0
}

func SetupPlaygroundConfig(dirname string) error {
	cmd := exec.Command("npx", "-y", "tsc", "--init")
	cmd.Dir = dirname
	out, err := cmd.Output()

	if err != nil {
		fmt.Fprint(os.Stderr, string(out))

		if exitErr, isExitErr := err.(*exec.ExitError); isExitErr {
			fmt.Fprint(os.Stderr, string(exitErr.Stderr))
		}

		return errors.Join(
			errors.New("failed to setup playground"),
			err,
		)
	}

	return nil
}

type PlaygroundFiles struct {
	File    *os.File
	DirName string
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

	tempdir, err := os.MkdirTemp(osTempDir, "ts-playground-*")
	if err != nil {
		return nil, errors.Join(
			errors.New("failed to create temporary directory"),
			err,
		)
	}
	out := PlaygroundFiles{
		DirName: tempdir,
	}

	tempfileName := path.Join(tempdir, "playground.ts")
	tempfile, err := os.Create(tempfileName)
	if err != nil {
		CleanupFiles(&out)
		return nil, errors.Join(
			errors.New("failed to create temporary file"),
			err,
		)
	}

	return &PlaygroundFiles{
		File:    tempfile,
		DirName: tempdir,
	}, nil
}
