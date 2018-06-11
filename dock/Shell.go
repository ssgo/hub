package dock

import (
	"os/exec"
	"log"
	"strings"
)

var shellFunc func(nodeName string, args ...string) (out string)

func SetShell(shell func(nodeName string, args ...string) (out string)) {
	shellFunc = shell
}

func defaultShell(nodeName string, args ...string) string {
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
	bytes, err := cmd.Output()

	if err != nil {
		log.Print("Dock	exec error	ssh ", strings.Join(sshArgs, " "), err.Error(), "	times: ", nodeFailedTimes[nodeName])
		nodeFailedTimes[nodeName] ++
		//if nodeFailedTimes[nodeName] >= 5 {
		//	log.Print("Dock	remove bad Node	", nodeName)
		//cc.HDEL("_nodes", nodeName)
		//}
		return ""
	}
	nodeFailedTimes[nodeName] = 0
	result := string(bytes)
	return result
}
