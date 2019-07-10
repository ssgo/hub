var GateConfigAction = {
    'getGateway': function (ctx) {
        ctx.http.get('/gateway').then(function (data) {
            ctx.states.set('gatewayConfig', data)
            ctx.resolve()
        }).catch(ctx.reject)
    },
    'setGateway': function(ctx, gatewayConfig) {
        ctx.http.post('/gateway', gatewayConfig).then(function (data) {
            ctx.states.set('status', data)
            ctx.resolve()
        }).catch(ctx.reject)
    }
}
