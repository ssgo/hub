var LoginView = {
    html: 'views/Login.html',
    login: function () {
        actions.call('user.login', {accessToken: sha1('SSGO-' + this.data.token + '-Dock')}).catch(function (err) {
            this.setData({error: err})
        }.bind(this))
    }
}
