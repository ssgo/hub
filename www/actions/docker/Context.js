var ContextAction = {
    'getContexts': function (ctx) {
        ctx.http.get('/contexts').then(function (data) {
            ctx.states.set({
                contexts: data
            })
            ctx.resolve()
        }).catch(ctx.reject)
    },

    'getContext': function (ctx, args) {
        ctx.http.get('/' + args.name).then(function (data) {
            if(!data) data = {}
            ctx.states.set('ctx_' + args.name, data)
            ctx.resolve()
        }).catch(ctx.reject)
    },

    'getStatus': function (ctx, args) {
        ctx.http.get('/' + args.name + '/status').then(function (data) {
            ctx.states.set('status_' + args.name, data)
            ctx.resolve()
        }).catch(ctx.reject)
    },

    'save': function (ctx, args) {
        ctx.http.post('/' + args.name, args).then(function (data) {
            ctx.resolve(data)
        }).catch(ctx.reject)
    },

    'remove': function (ctx, args) {
        ctx.http.delete('/' + args.name).then(function (data) {
            ctx.resolve(data)
        }).catch(ctx.reject)
    }
}
