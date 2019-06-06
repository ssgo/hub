var DockerView = {
    html: 'views/docker/Docker.html',
    stateBinds: ['contexts', 'authLevel', 'editMode'],

    getSubView: function (subName) {
        if (subName === 'global') {
            return new GlobalView()
        } else {
            return new ContextView(subName)
        }
    },

    onShow: function (path, nextPath) {
        if (nextPath) {
            this.data.nav = nextPath.name
        }
        route.bind('docker.*', this)
        actions.call('context.getContexts')
    },

    onHide: function () {
        route.unbind('docker.*', this)
    },

    onRoute: function (data) {
        this.setData({nav: data.last.name})
    },

    newContext: function () {
        var name = prompt('Enter a context name')
        if (name && /^[A-Za-z0-9_\.]+$/.test(name)) {
            this.data.contexts[name] = ''
            var that = this
            actions.call('context.save', {
                name: name,
                desc: '',
                apps: {},
                vars: {},
                binds: {}
            }).then(function () {
                that.setData({contexts: that.data.contexts}).then(function () {
                    route.go('/docker/' + name)
                })
            }).catch(function (reason) {
                alert('Create context has error: ' + reason)
            })
        } else {
            alert('Context name is require [A-Za-z0-9_\\.]+')
        }
    },

    // navTo: function (path) {
    //     // this.setData({nav: path})
    //     route.go('/docker/' + path)
    // }
}
