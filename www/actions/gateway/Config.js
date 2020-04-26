var GateConfigAction = {
    'getGateway': function (ctx) {
        ctx.http.get('/gateway').then(function (data) {
            if (data.configs) {
                ctx.states.set('gatewayConfig', data.configs)
                ctx.resolve()
            } else {
                ctx.reject('failed')
            }
        }).catch(ctx.reject)
    },
    'setGateway': function (ctx, gatewayConfig) {
        ctx.http.post('/gateway', {configs: gatewayConfig}).then(function (result) {
            if (result) {
                ctx.resolve()
            } else {
                ctx.reject('failed')
            }
        }).catch(ctx.reject)
    }
}
