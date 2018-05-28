var NodesAction = {
    'list': function (ctx) {
        ctx.http.get('/nodes').then(function (data) {
            ctx.states.set('nodes', data)
            ctx.resolve()
        }).catch(ctx.reject)
    },

    'getStatus': function (ctx) {
        ctx.http.get('/nodes/status').then(function (data) {
            ctx.states.set({
                nodeStatus: data
            })
            ctx.resolve()
        }).catch(ctx.reject)
    },

    'save': function (ctx, args) {
        ctx.http.post('/nodes', {nodes: args.nodes}).then(function () {
            ctx.resolve()
        }).catch(ctx.reject)
    }
}
