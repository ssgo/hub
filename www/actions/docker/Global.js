var GlobalAction = {
    'list': function (ctx) {
        ctx.http.get('/global').then(function (data) {
            ctx.states.set('global', data)
            ctx.resolve()
        }).catch(ctx.reject)
    },

    'getStatus': function (ctx) {
        ctx.http.get('/global/status').then(function (data) {
            ctx.states.set({
                nodeStatus: data.nodes,
                // registryStatus: data.registryStatus
            })
            ctx.resolve()
        }).catch(ctx.reject)
    },

    'save': function (ctx, args) {
        ctx.http.post('/global', args).then(function () {
            ctx.resolve()
        }).catch(ctx.reject)
    }
}
