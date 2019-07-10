var GatewayView = {
    html: 'views/gateway/Gateway.html',
    stateBinds: ['authLevel', 'editMode'],
    getSubView: function (subName) {
        return new GateConfigView()
    },

    onShow: function (path, nextPath) {
        if (nextPath) {
            this.data.nav = nextPath.name
        }
        route.bind('gate.*', this)
    },

    onHide: function () {
        route.unbind('gate.*', this)
    },

    onRoute: function (data) {
        this.setData({nav: data.last.name})
    }
}