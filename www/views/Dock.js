var DockView = {
    html: 'views/Dock.html',
    stateBinds: ['contexts', 'authLevel', 'editMode'],

    getSubView: function (subName) {
        if (subName === 'nodes') {
            return new NodesView()
        } else {
            return new ContextView(subName)
        }
    },

    onShow: function (path, nextPath) {
        if (nextPath) {
            this.data.nav = nextPath.name
        }
        route.bind('dock.*', this)
        actions.call('context.getContexts')
    },

    onHide: function () {
        route.unbind('dock.*', this)
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
                    route.go('/dock/' + name)
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
    //     route.go('/dock/' + path)
    // }
}
