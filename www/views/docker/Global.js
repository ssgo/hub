var GlobalView = function () {
    this.html = 'views/docker/Global.html'
    this.stateBinds = ['authLevel','editMode']
    this.stateRegisters = {global: [this, 'setGlobalData']}
    this.isActive = false
    this.data = {host: location.host, protocol: location.protocol}
    // this.refreshTid = 0
}

GlobalView.prototype.onShow = function () {
    var that = this
    if (this.data.authLevel === 2 && states.state.editMode === true){
        states.set({editMode:false})
    }
    actions.call('global.list').then(function () {
        setTimeout(that.refreshStatus, 100, that)
    })
    this.isActive = true
    states.state.currentModule = this
    // this.refreshTid = setInterval(this.refreshStatus, 5000, this)
}

GlobalView.prototype.canHide = function () {
    if (this.data.changed) {
        if (!confirm('Data has changed, do you want drop them?')) return false
        this.data.changed = false
    }
    return true
}

GlobalView.prototype.onHide = function () {
    this.isActive = false
    // clearInterval(this.refreshTid)
    // this.refreshTid = 0
}

GlobalView.prototype.setGlobalData = function (data) {
    if (data && data.global) {
        data = data.global
        var nodes = []
        for (var k in data.nodes) {
            data.nodes[k].name = k
            nodes.push(data.nodes[k])
        }

        var vars = []
        for (var k in data.vars) {
            vars.push({name: k, value: data.vars[k]})
        }

        var _nodes = CP(nodes)
        var _vars = CP(vars)

        nodes.push({})
        vars.push({})

        this.setData({
            nodes: nodes,
            vars: vars,
            _nodes: _nodes,
            _vars: _vars,
            args: data.args,
            publicKey: data.publicKey,
            installToken: data.installToken,
        })
    }
}

GlobalView.prototype.refreshStatus = function (that) {
    actions.call('global.getStatus').then(function () {
        that.onRefreshStatus()
    })
}

GlobalView.prototype.onRefreshStatus = function () {
    for (var k in this.data.nodes) {
        var node = this.data.nodes[k]
        if (!node.name) continue
        var status = states.state.nodeStatus[node.name]
        if (typeof status === "undefined") {
            continue
        }
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

GlobalView.prototype.save = function () {
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
        nodes[node.name.trim()] = {cpu: cpu, memory: memory}
    }

    var vars = {}
    for (var k in this.data.vars) {
        var v = this.data.vars[k]
        if (!v.name) {
            continue
        }
        vars[v.name.trim()] = v.value
    }

    var that = this
    actions.call('global.save', {nodes: nodes, vars: vars, args: this.data.args}).then(function () {
        that.setData({changed: false})
        that.onShow()
    }).catch(function (reason) {
        alert('Save global has error: ' + reason)
    })
}

GlobalView.prototype.check = function (event, type, idx) {
    var oldList = this.data['_' + type]
    var list = this.data[type]
    if ((idx < oldList.length && JSON.stringify(list[idx]) !== JSON.stringify(oldList[idx])) ||
        (idx >= oldList.length && list[idx].name)) {
        list[idx].changed = true
        if (this.data.changed !== true) {
            this.data.changed = true
        }
        // tpl.refresh(event.target.parentElement.parentElement, {index: idx, item: list[idx]})
        // this.refreshView()
    }
    if (idx == list.length - 1) {
        list.push({})
        this.refreshView()
    }else{
        tpl.refresh(this.$('.saveBox'), {data:{changed:true}})
        event.target.parentElement.parentElement.className = 'danger'
    }
}
