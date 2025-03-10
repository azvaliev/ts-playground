package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
)

func CreatePlayground() (*Playground, error) {
	playground, err := setupTempFiles()
	if err != nil {
		return playground, err
	}

	fmt.Printf("Created temp file at %s\n", playground.File.Name())

	err = setupPlaygroundConfig(playground.DirName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return playground, err
	}

	fmt.Printf("Setup playground at %s\n", playground.DirName)

	return playground, nil
}

type Playground struct {
	File    *os.File
	DirName string
}

func (playground *Playground) GetRelativeFilename() string {
	if playground.File == nil {
		return ""
	}

	_, filename := path.Split(playground.File.Name())
	return filename
}

func (playground *Playground) Destroy() error {
	if playground == nil {
		return nil
	}

	var err error
	if playground.File != nil {
		_ = playground.File.Close()
	}

	if playground.DirName != "" {
		err = os.RemoveAll(playground.DirName)
	}

	return err
}

func setupPlaygroundConfig(dirname string) error {
	// create tsconfig
	{
		cmd := exec.Command("npx", "-y", "tsc", "--init")
		cmd.Dir = dirname

		if out, err := cmd.Output(); err != nil {
			fmt.Println(string(out))
			return handleExecError(err, "Failed to setup playground project")
		}
	}

	return nil
}

func setupTempFiles() (*Playground, error) {
	osTempDir := os.TempDir()
	if osTempDir == "" {
		return nil, errors.New("failed to retrieve temporary directory")
	}

	out := Playground{}

	var err error
	out.DirName, err = os.MkdirTemp(osTempDir, "ts-playground-*")
	if err != nil {
		return nil, errors.Join(
			errors.New("failed to create temporary directory"),
			err,
		)
	}

	tempfileName := path.Join(out.DirName, "playground.js")
	out.File, err = os.Create(tempfileName)
	if err != nil {
		out.Destroy()
		return nil, errors.Join(
			errors.New("failed to create temporary file"),
			err,
		)
	}

	err = os.Chmod(out.File.Name(), os.ModePerm)
	if err != nil {
		out.Destroy()
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
