package dock

import (
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
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
	//if hubConfig.PrivateKey != "" {
	//	sshArgs = append(sshArgs, "-i", "/opt/privateKey", "-o", "StrictHostKeyChecking=no")
	//}
	sshArgs = append(sshArgs, "-i", dataPath(".ssh", "id_ecdsa"), "-o", "StrictHostKeyChecking=no")
	sshArgs = append(sshArgs, "docker@"+sshHost, "-p", sshPort, "docker "+strings.Join(args, " "))
	//sshArgs = append(sshArgs, args...)
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
				err = cmd.Start()
				if err == nil {
					bytes, err = ioutil.ReadAll(op)
					if err == nil {
						errorBytes, _ = ioutil.ReadAll(ep)
					}
					err = cmd.Wait()
				}
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
			_ = cmd.Process.Kill()
		}
		err = errors.New("timeout")
	}
	//endTime := time.Now()
	usedTime := int(time.Duration(time.Now().UnixNano()-startTime.UnixNano()) / time.Millisecond)

	if errorBytes != nil && len(errorBytes) > 0 {
		errStr := string(errorBytes)
		if !strings.Contains(errStr, "known hosts") && !strings.Contains(errStr, "Downloaded newer image") {
			err = errors.New(errStr)
		}
	}

	if err != nil {
		logError("docker exec failed: "+err.Error(),
			"node", nodeName,
			"command", args[0],
			"shell", "ssh "+strings.Join(sshArgs, " "),
			"failedTimes", nodeFailedTimes[nodeName],
			"usedTime", usedTime,
			"limitTime", timeout,
		)
		//log.Print("Dock	exec error	ssh ", strings.Join(sshArgs, " "),	"	error: ", err.Error(), "	times: ", nodeFailedTimes[nodeName], "	Used: ", time.Duration(endTime.UnixNano() - startTime.UnixNano()), " of ", timeout * time.Millisecond)
		nodeFailedTimes[nodeName]++
		return "", usedTime, err
	}
	nodeFailedTimes[nodeName] = 0
	result := string(bytes)
	return result, usedTime, nil
}
