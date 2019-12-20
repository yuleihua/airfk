package cmd

import (
	"os"
	"os/exec"
)

func ExecWithWait(cmdfile, conf string) error {
	var bin *exec.Cmd
	if conf != "" {
		bin = exec.Command(cmdfile, "-c", conf)
	} else {
		bin = exec.Command("sh", "-c", cmdfile)
	}
	bin.Stderr = os.Stderr
	bin.Stdout = os.Stdout

	if err := bin.Start(); err != nil {
		return err
	}

	chanError := make(chan error, 1)
	go func(chanErr chan error) {
		err := bin.Wait()
		if err != nil {
			//if exitError, ok := err.(*exec.ExitError); ok {
			//	processState := exitError.ProcessState
			//}
			chanErr <- err
		}
		chanErr <- nil
	}(chanError)

	err := <-chanError
	return err
}
