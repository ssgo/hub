var ContextView = function (name) {
    this.name = name
    this.html = 'views/Context.html'
    this.stateBinds = ['authLevel', 'editMode']
    this.stateRegisters = {}
    this.stateRegisters['ctx_' + this.name] = [this, 'setContextData']
    this.isActive = false
    // this.refreshTid = 0
}

ContextView.prototype.onShow = function () {
    var that = this
    actions.call('context.getContext', {name: this.name}).then(function () {
        setTimeout(that.refreshStatus, 100, that)
    })
    this.isActive = true
    states.state.currentModule = this
    // this.refreshTid = setInterval(this.refreshStatus, 5000, this)
}

ContextView.prototype.canHide = function () {
    if (this.data.changed) {
        if (!confirm('Data has changed, do you want drop them?')) return false
        this.data.changed = false
    }
    return true
}

ContextView.prototype.onHide = function () {
    this.isActive = false
    // clearInterval(this.refreshTid)
    // this.refreshTid = 0
}

ContextView.prototype.setContextData = function (data) {
    if (data && data['ctx_' + this.name]) {
        data = data['ctx_' + this.name]

        var vars = []
        for (var k in data.vars) {
            vars.push({name: k, value: data.vars[k]})
        }

        var binds = []
        for (var k in data.binds) {
            binds.push({name: k, value: data.binds[k]})
        }

        var apps = []
        for (var k in data.apps) {
            data.apps[k].name = k
            apps.push(data.apps[k])
        }

        var _vars = CP(vars)
        var _binds = CP(binds)
        var _apps = CP(apps)

        vars.push({})
        binds.push({})
        apps.push({})

        this.setData({
            name: data.name,
            desc: data.desc,
            vars: vars,
            binds: binds,
            apps: apps,
            _vars: _vars,
            _binds: _binds,
            _apps: _apps,
        })
    }
}


ContextView.prototype.refreshStatus = function (that) {
    actions.call('context.getStatus', {name: that.name}).then(function () {
        that.onRefreshStatus()
    })
}

ContextView.prototype.onRefreshStatus = function () {
    var status = states.state['status_' + this.name]
    for (var k in this.data.apps) {
        var app = this.data.apps[k]
        if (!app.name) continue
        var target = this.$('.' + 'status_box_' + k)
        if (!target) continue
        var runs = status[app.name]
        if (!runs || !(runs instanceof Array)) continue
        for (k2 in runs) {
            runs[k2].showUpTimeColor = runs[k2].upTime.indexOf('(healthy)') !== -1 ? '#090' : (runs[k2].upTime.indexOf('(') !== -1 ? '#f22' : '#00f')
            runs[k2].showUpTime = runs[k2].upTime
                .replace('About a', '1')
                .replace('Lessthan a', '0')
                .replace(/Up |econds|econd|inutes|inute|ours|our|ays|ay| /g, '')
                .replace(/\(/g, ' ')
                .replace(/\)/g, '')
                .replace(/healthy/g, '✓')
        }
        app.runs = runs
        tpl.refresh(target, {item: app, index: k})
    }
}

ContextView.prototype.save = function () {
    var apps = {}
    for (var k in this.data.apps) {
        var v = this.data.apps[k]
        if (!v.name) {
            continue
        }

        var cpu = parseInt(v.cpu)
        var memory = parseInt(v.memory)
        if (isNaN(cpu) || isNaN(memory) || cpu < 1 || cpu > 1024 || memory < 1 || memory > 10240) {
            alert('Cpu: ' + cpu + ' (1~1024) or Memory: ' + memory + ' (1~10240) is not available')
            return false
        }

        var min = parseInt(v.min)
        var max = parseInt(v.max)
        if (isNaN(min) || isNaN(max) || min < 1 || min > 1024 || max < 1 || max > 10240) {
            alert('Min: ' + min + ' (1~1024) or Max: ' + max + ' (1~10240) is not available')
            return false
        }
        if (min > max) {
            alert('Min: ' + min + ' > Max: ' + max + ' is not acceptable')
            return false
        }

        apps[v.name.trim()] = {
            desc: v.desc,
            cpu: cpu,
            memory: memory,
            min: min,
            max: max,
            args: v.args,
            command: v.command,
            memo: v.memo,
            active: v.active === true
        }
        // delete apps[v.name.trim()]['name']
    }

    var vars = {}
    for (var k in this.data.vars) {
        var v = this.data.vars[k]
        if (!v.name) {
            continue
        }
        vars[v.name.trim()] = v.value
    }

    var binds = {}
    for (var k in this.data.binds) {
        var v = this.data.binds[k]
        if (!v.name) {
            continue
        }
        binds[v.name.trim()] = v.value
    }

    var that = this
    actions.call('context.save', {
        name: this.name.trim(),
        desc: this.data.desc,
        apps: apps,
        vars: vars,
        binds: binds
    }).then(function (result) {
        if (result.ok) {
            that.setData({changed: false})
            that.onShow()
        } else {
            alert('Save context has failed, '+result.error)
        }
    }).catch(function (reason) {
        alert('Save context has error: ' + reason)
    })
}

ContextView.prototype.remove = function () {
    if (prompt('Please enter the context name to conform for remove') === this.name) {
        actions.call('context.remove', {name: this.name}).then(function () {
            route.go('/dock/global')
            actions.call('context.getContexts')
        }).catch(function (reason) {
            alert('Remove context has error: ' + reason)
        })
    }
}

ContextView.prototype.check = function (event, type, idx) {
    var oldList = this.data['_' + type]
    var list = this.data[type]
    var changed = false
    if ((idx < oldList.length && JSON.stringify(list[idx]) !== JSON.stringify(oldList[idx])) ||
        (idx >= oldList.length && list[idx].name)) {
        list[idx].changed = true
        if (this.data.changed !== true) {
            this.data.changed = true
        }
        // tpl.refresh(event.target.parentElement.parentElement, {index: idx, item: list[idx]})
        changed = true
    }
    if (idx === list.length - 1) {
        list.push({})
        changed = true
    }

    if (changed === true) {
        this.refreshView()
    }
}
