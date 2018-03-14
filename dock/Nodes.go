package dock

import (
	"strings"
	"strconv"
	"log"
	"fmt"
)

type NodeInfo struct {
	TotalCpu    float32
	TotalMemory float32
	UsedCpu     float32
	UsedMemory  float32
}

func (node *NodeInfo) String() string {
	return fmt.Sprintf("%.2f,%.2f", node.TotalCpu, node.TotalMemory)
}

func NewNode(nodeString string) *NodeInfo {
	a := strings.Split(nodeString, ",")
	if len(a) < 2 {
		return nil
	}
	cpu, err1 := strconv.ParseFloat(a[0], 10)
	memory, err2 := strconv.ParseFloat(a[1], 10)
	if err1 != nil || err2 != nil {
		return nil
	}
	return &NodeInfo{TotalCpu: float32(cpu), TotalMemory: float32(memory)}
}

var nodes = map[string]*NodeInfo{}

func updateNodesInfo() bool {
	changed := false

	// 使用配置中的节点
	remoteNodes := map[string]string{}
	for nodeName, nodeString := range config.Nodes {
		remoteNodes[nodeName] = *nodeString
	}
	dcCache.Do("HGETALL", "_nodes").To(&remoteNodes)

	for nodeName, nodeString := range remoteNodes {
		node := NewNode(nodeString)
		if node == nil {
			continue
		}

		if nodes[nodeName] != nil && nodes[nodeName].String() == node.String() {
			continue
		}

		// 失败超过5次的节点自动忽略
		if nodeFailedTimes[nodeName] >= 5 {
			continue
		}

		if nodes[nodeName] != nil {
			log.Printf("Dock	nodes	update	%s	%s => %s", nodeName, nodes[nodeName].String(), nodeString)
		} else {
			log.Printf("Dock	nodes	add	%s	%s", nodeName, nodeString)
		}
		changed = true
		nodes[nodeName] = node
	}

	for nodeName, node := range nodes {
		// 失败超过5次的节点自动忽略
		if remoteNodes[nodeName] == "" || nodeFailedTimes[nodeName] >= 5 {
			log.Printf("Dock	nodes	remove	%s	%s", nodeName, node)
			changed = true
			delete(nodes, nodeName)
		}
	}
	return changed
}

func nextMinScoreNode(app *AppInfo) string {
	var minScore float32 = -1
	minNodeName := ""
	for nodeName, node := range nodes {
		score := node.UsedMemory/node.TotalMemory + node.UsedCpu/node.TotalCpu
		for _, run := range app.Runs {
			// 已经有过的节点得分 +1，优先考虑平均分布
			if run.Node == nodeName {
				if strings.Index(app.Args, " -v ") != -1 || strings.Index(app.Args, " --volume ") != -1 {
					// 挂载磁盘的，尽可能的分布到不同节点，增加 10000% 权重
					score += 100
				} else {
					if app.Min <= 2 {
						// 2个节点 强平均分配，增加 300% 权重
						score += 3
					} else if app.Min <= 4 {
						// 3~4个节点 较强平均分配，增加 150% 权重
						score += 1.5
					} else if app.Min <= 6 {
						// 5~6个节点 略强平均分配，增加 80% 权重
						score += 0.8
					} else {
						// 7个及以上节点 弱平均分配，增加 30% 权重
						score += 0.3
					}
				}
			}
		}
		if minScore < 0 || score < minScore {
			minScore = score
			minNodeName = nodeName
		}
	}
	return minNodeName
}
