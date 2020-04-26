var GateConfigView = function () {
    this.html = 'views/gateway/Config.html'
    this.stateBinds = ['authLevel', 'editMode']
    this.stateRegisters = {}
    this.stateRegisters['gatewayConfig'] = [this, 'setGatewayConfig']
    this.isActive = false
    this.data = {}
}

GateConfigView.prototype.setGatewayConfig = function () {
    let configs = states.state.gatewayConfig
    if (!configs) configs = []
    var _configs = CP(configs)
    configs.push({})
    this.setData({
        configs: configs,
        _configs: _configs,
    })
}

GateConfigView.prototype.onShow = function () {
    var that = this
    if (this.data.authLevel === 2 && states.state.editMode === true) {
        states.set({editMode: false})
    }
    actions.call("gateway.getGateway")
    this.isActive = true
    states.state.currentModule = this
}

GateConfigView.prototype.canHide = function () {
    if (this.data.changed) {
        if (!confirm('Data has changed, do you want drop them?')) return false
        this.data.changed = false
    }
    return true
}

GateConfigView.prototype.onHide = function () {
    this.isActive = false
}

GateConfigView.prototype.check = function (event, type, idx) {
    let oldList = this.data['_' + type]
    let list = this.data[type]
    if ((idx < oldList.length && JSON.stringify(list[idx]) !== JSON.stringify(oldList[idx])) ||
        (idx >= oldList.length && list[idx].key)) {
        this.data[type][idx].changed = true
    }
    if (idx == list.length - 1) {
        list.push({})
        this.refreshView()
    } else {
        tpl.refresh(this.$('.saveBox'), {data: {changed: true}})
        event.target.parentElement.parentElement.className = 'danger'
        // event.target.parentElement.className = 'bg-danger'
    }
}

GateConfigView.prototype.save = function () {
    let list = []
    for (let data of this.data.configs) {
        if (data.key && data.field) {
            list.push(data)
        }
    }

    var that = this
    actions.call('gateway.setGateway', list).then(function () {
        that.setData({changed: false})
        that.onShow()
    }).catch(function (reason) {
        alert('Save gateway info has error: ' + reason.toString())
    })
}