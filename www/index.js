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
actions.register('global', GlobalAction)
actions.register('context', ContextAction)
actions.register('gateway', GateConfigAction)

// 设置根路由Root
route.Root = {
    getSubView: function (subName) {
        switch (subName) {
            case 'login':
                return LoginView
            case 'docker':
                return DockerView
            case 'gateway':
                return GatewayView
        }
    }
}

var startRoute = location.hash ? location.hash.substring(1) : ''
route.bindHash()
states.bind('logined', function (data) {
    if (data.logined) {
        var navs = $$('.navbar-nav > li')
        if (!startRoute || /^\/login/.test(startRoute)) {
            route.go('/docker/global')
        } else {
            route.go(startRoute)
        }

        if (route._prevPaths.length > 0) {
            setNavStatus($('.' + route._prevPaths[0].name + 'Nav'))
        }
    } else {
        route.go('/login')
    }
})

window.addEventListener('load', function () {
    actions.call('user.login', {accessToken: sessionStorage.accessToken})
})

setInterval(function () {
    var m = states.state.currentModule
    if( m && m['isActive'] && m['refreshStatus'] && typeof m['refreshStatus'] === 'function' ){
        m.refreshStatus(m)
    }
}, 5000, this)

function setNavStatus(target) {
    for (var li of $$('.topNav li')) {
        li.className = ''
    }
    if(!target) target = $('.navbar-nav > li')
    target.className = 'active'
}
