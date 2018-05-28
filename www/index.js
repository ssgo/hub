// 选择器
window.$ = function (selector) {
    if (!selector) return document.body
    return document.querySelector(selector)
}
window.$$ = function (selector) {
    if (!selector) return [document.body]
    return document.querySelectorAll(selector)
}
window.L = function (obj) {
    console.log(obj)
}
window.CP = function (data) {
    return JSON.parse(JSON.stringify(data))
}

var states = new svcState.State('binds')
var http = new svcWeb.Http('//' + location.host)
var route = new svcWeb.Route(states)
var tpl = new svcWeb.Tpl()
var actions = new svcAction.Action({
    states: states,
    route: route,
    http: http
})

actions.register('user', UserAction)
actions.register('nodes', NodesAction)
actions.register('context', ContextAction)

// 设置根路由Root
route.Root = {
    getSubView: function (subName) {
        switch (subName) {
            case 'login':
                return LoginView
            case 'dock':
                return DockView
        }
    }
}

var startRoute = location.hash ? location.hash.substring(1) : ''
route.bindHash()
states.bind('logined', function (data) {
    if (data.logined) {
        if (!startRoute || /^\/login/.test(startRoute)) {
            route.go('/dock/nodes')
        } else {
            route.go(startRoute)
        }
    } else {
        route.go('/login')
    }
})

window.addEventListener('load', function () {
    actions.call('user.login', {accessToken: sessionStorage.accessToken}).catch(function (reason) {
    })
})
