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
    // states.set({editMode: true})
    // setTimeout(function () {
    //     that.showConfigWindow(0)
    // }, 100)
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
            alert('Save context has failed, ' + result.error)
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
    if ((idx < oldList.length && JSON.stringify(list[idx]) !== JSON.stringify(oldList[idx])) ||
        (idx >= oldList.length && list[idx].name)) {
        list[idx].changed = true
        if (this.data.changed !== true) {
            this.data.changed = true
        }

        // sync bind
        var synced = false
        if (type==='apps' && list[idx].name && list[idx].name !== oldList[idx].name ){
            for (var i in this.data._binds){
                if (this.data._binds[i].name === oldList[idx].name ){
                    this.data.binds[i].name = list[idx].name
                    this.data.binds[i].changed = true
                    synced = true
                }
            }
        }

        if (type==='binds' && list[idx].name && list[idx].name !== oldList[idx].name ){
            for (var i in this.data._apps){
                if (this.data._apps[i].name === oldList[idx].name ){
                    this.data.apps[i].name = list[idx].name
                    this.data.apps[i].changed = true
                    synced = true
                }
            }
        }

        // tpl.refresh(event.target.parentElement.parentElement, {index: idx, item: list[idx]})
    }
    if (idx === list.length - 1) {
        list.push({})
        this.refreshView()
    } else if (synced){
        this.refreshView()
    }else{
        tpl.refresh(this.$('.saveBox'), {data: {changed: true}})
        event.target.parentElement.parentElement.className = 'danger'
    }
}

ContextView.prototype.checkConfig = function (type, idx) {
    var list = this.data[type]
    if (idx === list.length - 1) {
        list.push({})
        this.refreshView()
    }
}


ContextView.prototype.showConfigWindow = function (which, index) {
    if (!states.state.global) {
        var that = this
        actions.call('global.list').then(function () {
            that.showConfigWindow(which, index)
        })
        return
    }

    var data = {
        configWindowShowing: true,
        configIsHost: false,
        configGlobalVars: [],
        configContextVars: [],
        configPorts: [],
        configVolumes: [],
        configEnvs: [],
        configRefVars: [],
        configOthers: [],
    }

    for (var k in states.state.global.vars) {
        data.configGlobalVars.push({key: '${' + k + '}', value: states.state.global.vars[k]})
    }
    for (var k in this.data.vars) {
        var v = this.data.vars[k]
        if (v.name) {
            data.configContextVars.push({key: '${' + v.name + '}', value: v.value})
        }
    }

    var args = ''
    if (which==='app') {
        this.configAppItem = this.data.apps[index]
        this.configVarItem = null
        args = this.configAppItem.args
    }else if (which==='var') {
        this.configVarItem = this.data.vars[index]
        this.configAppItem = null
        args = this.configVarItem.value
    }
    var m = praseCommandArgs(args+' ')
    var num = m.length
    for (var i = 0; i < m.length; i++) {
        var v = m[i]
        if (v === '-e') {
            if (i < num - 1) {
                a = m[i + 1].split('=', 2)
                data.configEnvs.push({key: a[0], value: a[1]})
                i++
            }
        } else if (v === '-p') {
            if (i < num - 1) {
                a = m[i + 1].split(':', 2)
                data.configPorts.push({from: a[0], to: a[1]})
                i++
            }
        } else if (v === '-v') {
            if (i < num - 1) {
                a = m[i + 1].split(':', 2)
                data.configVolumes.push({from: a[0], to: a[1]})
                i++
            }
        } else if (v.length > 3 && v[0] === '$' && v[1] === '{' && v[v.length - 1] === '}') {
            data.configRefVars.push({key: v})
        } else if (v === '--network=host') {
            data.configIsHost = true
        } else {
            data.configOthers.push({value: v})
        }
    }
    data.configPorts.push({})
    data.configVolumes.push({})
    data.configEnvs.push({})
    data.configRefVars.push({})
    data.configOthers.push({})
    this.setData(data)
}

ContextView.prototype.showVarHinter = function (target, value) {
    if (value.length > 3 && value[0] === '$' && value[1] === '{' && value[value.length - 1] === '}') {
        var k = value.substring(2, value.length - 1)
        value = states.state.global.vars[k] || states.state['ctx_' + this.name].vars[k]
    }

    var m = praseCommandArgs(value)
    var num = m.length
    var hints = []
    for (var i = 0; i < m.length; i++) {
        var v = m[i]
        if (v === '-e') {
            if (i < num - 1) {
                a = m[i + 1].split('=', 2)
                hints.push('-e <b>' + a[0] + '</b>=<i>' + a[1] + '</i>')
                i++
            }
        } else if (v === '-p') {
            if (i < num - 1) {
                a = m[i + 1].split(':', 2)
                hints.push('-p <b>' + a[0] + '</b>:<i>' + a[1] + '</i>')
                i++
            }
        } else if (v === '-v') {
            if (i < num - 1) {
                a = m[i + 1].split(':', 2)
                hints.push('-v <b>' + a[0] + '</b>:<i>' + a[1] + '</i>')
                i++
            }
        } else if (v.length > 3 && v[0] === '$' && v[1] === '{' && v[v.length - 1] === '}') {
            hints.push('<i>' + v + '</i>')
        } else {
            hints.push('<i>' + v + '</i>')
        }
    }

    var o = this.$('.varHinter')
    o.innerHTML = hints.join('<br/>')
    var x = getElementLeft(target)
    var y = getElementTop(target) + target.offsetHeight
    if (target.nodeName === 'INPUT') {
        y -= this.$('.configView').scrollTop
    }
    o.style.left = x + 'px'
    o.style.top = y + 'px'
    o.style.display = 'block'
}

ContextView.prototype.hideVarHinter = function (target, value) {
    var o = this.$('.varHinter')
    o.innerHTML = ''
    o.style.display = 'none'
}

function getElementLeft(element) {
    var actualLeft = element.offsetLeft;
    var current = element.offsetParent;

    while (current !== null) {
        actualLeft += current.offsetLeft;
        current = current.offsetParent;
    }

    return actualLeft;
}

function getElementTop(element) {
    var actualTop = element.offsetTop;
    var current = element.offsetParent;

    while (current !== null) {
        actualTop += current.offsetTop;
        current = current.offsetParent;
    }

    return actualTop;
}

ContextView.prototype.saveConfig = function () {
    var cfgs = []

    if (this.data.configRefVars.length > 0) {
        for (var d of this.data.configRefVars) {
            if (d.key) {
                cfgs.push(d.key)
            }
        }
    }

    if (this.data.configIsHost) {
        cfgs.push('--network=host')
    } else if (this.data.configPorts.length > 0) {
        for (var d of this.data.configPorts) {
            if (d.from && d.to) {
                cfgs.push('-p ' + d.from + ':' + d.to)
            }
        }
    }

    if (this.data.configEnvs.length > 0) {
        for (var d of this.data.configEnvs) {
            if (d.key) {
                var v = d.key + '=' + d.value
                if (v.indexOf(' ') !== -1 && v[0] !== '"' && v[0] !== "'") {
                    v = "'" + v.replace(/'/g, "\\'") + "'"
                }
                cfgs.push('-e ' + v)
            }
        }
    }

    if (this.data.configVolumes.length > 0) {
        for (var d of this.data.configVolumes) {
            if (d.from && d.to) {
                var v = d.from + ':' + d.to
                if (v.indexOf(' ') !== -1 && v[0] !== '"' && v[0] !== "'") {
                    v = "'" + v.replace(/'/g, "\\'") + "'"
                }
                cfgs.push('-v ' + v)
            }
        }
    }

    if (this.data.configOthers.length > 0) {
        for (var d of this.data.configOthers) {
            if (d.value) {
                cfgs.push(d.value)
            }
        }
    }

    if (this.configAppItem) {
        this.configAppItem.args = cfgs.join(' ')
        this.configAppItem.changed = true
    }else if (this.configVarItem) {
        this.configVarItem.value = cfgs.join(' ')
        this.configVarItem.changed = true
    }

    this.setData({
        changed: true,
        configWindowShowing: false,
    })
}

ContextView.prototype.hideConfigWindow = function () {
    this.setData({
        configWindowShowing: false
    })
}

function praseCommandArgs(cmd) {
    if (!cmd) {
        return []
    }
    cmd = cmd.trim() + ' '
    var args = []
    var start = -1
    var quota = null
    for (var i = 0; i < cmd.length; i++) {
        var c = cmd[i]
        if (start === -1) {
            start = i
            if (c === '"' || c === '\'') {
                quota = c
            }
        } else if (c === ' ') {
            if (quota === null) {
                if (cmd[start] === cmd[i - 1] && (cmd[start] === '"' || cmd[start] === '\'')) {
                    args.push(cmd.substring(start + 1, i - 1).replace(new RegExp("\\\\'", 'g'), cmd[start]))
                } else {
                    args.push(cmd.substring(start, i))
                }
                start = -1
            }
        } else if (c === quota) {
            if (i > 0 && cmd[i - 1] !== '\\') {
                quota = null
            }
        }
    }
    return args
}
