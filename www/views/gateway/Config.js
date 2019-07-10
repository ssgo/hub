var GateConfigView = function () {
    this.html = 'views/gateway/Config.html'
    this.stateBinds = ['authLevel','editMode']
    this.isActive = false
    this.data = {host: location.host, protocol: location.protocol}
}

GateConfigView.prototype.getConfig = function() {
    var that = this
    actions.call("gateway.getGateway").then(function () {
        if(typeof states.state.gatewayConfig ==="undefined") {
            return
        }
        if(!states.state.gatewayConfig) {
            return
        }
        var gatewayInfo = {}
        var gatewayConfig = states.state.gatewayConfig

        var proxies = [];
        for(var key in gatewayConfig.proxies) {
            proxies.push({path:key, app:gatewayConfig.proxies[key]})
        }

        var rewrites = [];
        for(var key in gatewayConfig.rewrites) {
            rewrites.push({fromPath:key, toPath:gatewayConfig.rewrites[key]})
        }

        var _proxies = CP(proxies)
        var _rewrites = CP(rewrites)

        proxies.push({})
        rewrites.push({})

        that.setData({
            proxies:proxies,
            rewrites:rewrites,
            _proxies:_proxies,
            _rewrites:_rewrites,
            prefix:gatewayConfig.prefix
        })

    })
}

GateConfigView.prototype.onShow = function () {
    var that = this
    if (this.data.authLevel === 2 && states.state.editMode === true){
        states.set({editMode:false})
    }
    that.getConfig()
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
    var oldList = this.data['_' + type]
    var list = this.data[type]
    if ((idx < oldList.length && JSON.stringify(list[idx]) !== JSON.stringify(oldList[idx])) ||
        (idx >= oldList.length && list[idx].name)) {
        list[idx].changed = true
        if (this.data.changed !== true) {
            this.data.changed = true
        }
    }
    if (idx == list.length - 1) {
        list.push({})
        this.refreshView()
    }else{
        tpl.refresh(this.$('.saveBox'), {data:{changed:true}})
        event.target.parentElement.className = 'bg-danger'
    }
}

GateConfigView.prototype.save = function() {
    var proxies = {}
    for (var k in this.data.proxies) {
        var proxy = this.data.proxies[k]
        if (!proxy.path) {
            continue
        }
        var path = proxy.path.trim()
        var app = proxy.app.trim()
        if (path.length<1 || app.length<1) {
           continue
        }
        proxies[path] = app
    }

    var rewrites = {}
    for (var k in this.data.rewrites) {
        var rewrite = this.data.rewrites[k]
        if (!rewrite.fromPath) {
            continue
        }
        var fromPath = rewrite.fromPath.trim()
        var toPath = rewrite.toPath.trim()
        if (fromPath.length<1 || toPath.length<1) {
            continue
        }
        rewrites[fromPath] = toPath
    }

    var that = this
    actions.call('gateway.setGateway', {proxies: proxies, rewrites: rewrites}).then(function () {
        that.setData({changed: false})
        that.onShow()
    }).catch(function (reason) {
        alert('Save gateway info has error: ' + reason.toString())
    })
}