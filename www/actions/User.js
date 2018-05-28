var UserAction = {

    'login': function (ctx, args) {
        if (!args.accessToken) {
            ctx.states.set('logined', false)
            ctx.reject('No Token')
            return
        }

        ctx.http.upHeaders['Access-Token'] = args.accessToken
        ctx.http.post('/login').then(function (data) {
            if (data > 0) {
                ctx.states.set('logined', true)
                ctx.states.set('authLevel', data)
                sessionStorage.accessToken = args.accessToken
                ctx.resolve()
            } else {
                ctx.states.set('logined', false)
                ctx.reject(data || 'Bad Token')
            }
        }).catch(function (err) {
            ctx.states.set('logined', false)
            ctx.reject(err)
        })
    },

    'logout': function (ctx) {
        delete sessionStorage.accessToken
        ctx.states.set('logined', false)
    }

}
