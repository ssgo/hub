package dock

import (
	"os/exec"
	"strings"
	"time"
	"errors"
	"io"
	"io/ioutil"
	"github.com/ssgo/s"
)

var shellFunc func(timeout time.Duration, nodeName string, args ...string) (string, int, error)

func SetShell(shell func(timeout time.Duration, nodeName string, args ...string) (string, int, error)) {
	shellFunc = shell
}

func defaultShell(timeout time.Duration, nodeName string, args ...string) (string, int, error) {
	var sshHost, sshPort string
	{
		a := strings.Split(nodeName, ":")
		sshHost = a[0]
		if len(a) > 1 {
			sshPort = a[1]
		} else {
			sshPort = "22"
		}
	}
	sshArgs := make([]string, 0)
	if config.PrivateKey != "" {
		sshArgs = append(sshArgs, "-i", "/opt/privateKey", "-o", "StrictHostKeyChecking=no")
	}
	sshArgs = append(sshArgs, "docker@"+sshHost, "-p", sshPort, "docker")
	sshArgs = append(sshArgs, args...)
	cmd := exec.Command("ssh", sshArgs...)
	var bytes, errorBytes []byte
	var err error

	isOk := make(chan bool, 1)
	isTimeout := make(chan bool, 1)
	startTime := time.Now()
	go func() {
		//bytes, err = cmd.Output()
		var op, ep io.ReadCloser
		op, err = cmd.StdoutPipe()
		if err == nil {
			ep, err = cmd.StderrPipe()
			if err == nil {
				cmd.Start()
				bytes, err = ioutil.ReadAll(op)
				errorBytes, _ = ioutil.ReadAll(ep)
				cmd.Wait()
			}
		}
		isOk <- true
	}()
	go func() {
		time.Sleep(timeout * time.Millisecond)
		isTimeout <- true
	}()

	select {
	case <-isOk:
	case <-isTimeout:
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		err = errors.New("timeout")
	}
	//endTime := time.Now()
	usedTime := int(time.Duration(time.Now().UnixNano() - startTime.UnixNano()) / time.Millisecond)

	if errorBytes != nil && len(errorBytes) > 0 {
		errStr := string(errorBytes)
		if !strings.Contains(errStr, "known hosts") && !strings.Contains(errStr, "Downloaded newer image") {
			err = errors.New(errStr)
		}
	}

	if err != nil {
		s.Error("Dock", s.Map{
			"type":        "execFailed",
			"node":        nodeName,
			"command":     args[0],
			"shell":       "ssh " + strings.Join(sshArgs, " "),
			"failedTimes": nodeFailedTimes[nodeName],
			"usedTime":    usedTime,
			"limitTime":   timeout,
			"error":       err.Error(),
		})
		//log.Print("Dock	exec error	ssh ", strings.Join(sshArgs, " "),	"	error: ", err.Error(), "	times: ", nodeFailedTimes[nodeName], "	Used: ", time.Duration(endTime.UnixNano() - startTime.UnixNano()), " of ", timeout * time.Millisecond)
		nodeFailedTimes[nodeName] ++
		return "", usedTime, err
	}
	nodeFailedTimes[nodeName] = 0
	result := string(bytes)
	return result, usedTime, nil
}
