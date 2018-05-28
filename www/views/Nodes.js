var NodesView = function () {
    this.html = 'views/Nodes.html'
    this.stateBinds = ['authLevel','editMode']
    this.stateRegisters = {nodes: [this, 'setNodesData']}
    this.refreshTid = 0
}

NodesView.prototype.onShow = function () {
    var that = this
    actions.call('nodes.list').then(function () {
        setTimeout(that.refreshStatus, 100, that)
    })
    this.refreshTid = setInterval(this.refreshStatus, 5000, this)
}

NodesView.prototype.canHide = function () {
    if (this.data.changed) {
        if (!confirm('Data has changed, do you want drop them?')) return false
        this.data.changed = false
    }
    return true
}

NodesView.prototype.onHide = function () {
    clearInterval(this.refreshTid)
    this.refreshTid = 0
}

NodesView.prototype.setNodesData = function (data) {
    if (data && data.nodes) {
        data = data.nodes
        var nodes = []
        for (var k in data) {
            data[k].name = k
            nodes.push(data[k])
        }
        var _nodes = CP(nodes)
        if (states.state.authLevel >= 2) {
            nodes.push({})
        }

        this.setData({
            nodes: nodes,
            _nodes: _nodes
        })
    }
}

NodesView.prototype.refreshStatus = function (that) {
    actions.call('nodes.getStatus').then(function () {
        that.onRefreshStatus()
    })
}

NodesView.prototype.onRefreshStatus = function () {
    for (var k in this.data.nodes) {
        var node = this.data.nodes[k]
        if (!node.name) continue
        var status = states.state.nodeStatus[node.name]
        if (node.usedCpu !== status.usedCpu) {
            node.usedCpu = status.usedCpu
            tpl.refresh(this.$('.' + 'usedCpu_' + k), {item: node})
        }
        if (node.usedMemory !== status.usedMemory) {
            node.usedMemory = status.usedMemory
            tpl.refresh(this.$('.' + 'usedMemory_' + k), {item: node})
        }
        if (node.totalRuns !== status.totalRuns) {
            node.totalRuns = status.totalRuns
            tpl.refresh(this.$('.' + 'totalRuns_' + k), {item: node})
        }
    }
}

NodesView.prototype.save = function () {
    var nodes = {}
    for (var k in this.data.nodes) {
        var node = this.data.nodes[k]
        if (!node.name) {
            continue
        }
        var cpu = parseInt(node.cpu)
        var memory = parseInt(node.memory)
        if (isNaN(cpu) || isNaN(memory) || cpu < 1 || cpu > 1024 || memory < 1 || memory > 10240) {
            alert('Cpu: ' + cpu + ' (1~1024) or Memory: ' + memory + ' (1~10240) is not available')
            return false
        }
        nodes[node.name] = {cpu: cpu, memory: memory}
    }

    var that = this
    actions.call('nodes.save', {nodes: nodes}).then(function () {
        that.setData({changed: false})
        that.onShow()
    }).catch(function (reason) {
        alert('Save nodes has error: ' + reason)
    })
}

NodesView.prototype.check = function (event, type, idx) {
    var oldList = this.data['_' + type]
    var list = this.data[type]
    if ((idx < oldList.length && JSON.stringify(list[idx]) !== JSON.stringify(oldList[idx])) ||
        (idx >= oldList.length && list[idx].name)) {
        list[idx].changed = true
        if (this.data.changed !== true) {
            this.data.changed = true
        }
        // tpl.refresh(event.target.parentElement.parentElement, {index: idx, item: list[idx]})
        this.refreshView()
    }
    if (idx == list.length - 1) {
        list.push({})
        this.refreshView()
    }
}
