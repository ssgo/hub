(function (global, factory) {
	typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports) :
	typeof define === 'function' && define.amd ? define(['exports'], factory) :
	(factory((global.svcWeb = global.svcWeb || {})));
}(this, (function (exports) { 'use strict';

var _typeof = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? function (obj) {
  return typeof obj;
} : function (obj) {
  return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj;
};











var classCallCheck = function (instance, Constructor) {
  if (!(instance instanceof Constructor)) {
    throw new TypeError("Cannot call a class as a function");
  }
};

var createClass = function () {
  function defineProperties(target, props) {
    for (var i = 0; i < props.length; i++) {
      var descriptor = props[i];
      descriptor.enumerable = descriptor.enumerable || false;
      descriptor.configurable = true;
      if ("value" in descriptor) descriptor.writable = true;
      Object.defineProperty(target, descriptor.key, descriptor);
    }
  }

  return function (Constructor, protoProps, staticProps) {
    if (protoProps) defineProperties(Constructor.prototype, protoProps);
    if (staticProps) defineProperties(Constructor, staticProps);
    return Constructor;
  };
}();









var inherits = function (subClass, superClass) {
  if (typeof superClass !== "function" && superClass !== null) {
    throw new TypeError("Super expression must either be null or a function, not " + typeof superClass);
  }

  subClass.prototype = Object.create(superClass && superClass.prototype, {
    constructor: {
      value: subClass,
      enumerable: false,
      writable: true,
      configurable: true
    }
  });
  if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass;
};











var possibleConstructorReturn = function (self, call) {
  if (!self) {
    throw new ReferenceError("this hasn't been initialised - super() hasn't been called");
  }

  return call && (typeof call === "object" || typeof call === "function") ? call : self;
};

/*
global XMLHttpRequest
global location
 */

var _class = function () {
  function _class(url) {
    classCallCheck(this, _class);

    this.baseUrl = url;
    this.upHeaders = {};
    this.transmitHeaders = [];
  }

  createClass(_class, [{
    key: 'get',
    value: function get$$1(url) {
      return this.do('GET', url);
    }
  }, {
    key: 'post',
    value: function post(url, data) {
      return this.do('POST', url, data);
    }
  }, {
    key: 'put',
    value: function put(url, data) {
      return this.do('PUT', url, data);
    }
  }, {
    key: 'delete',
    value: function _delete(url, data) {
      return this.do('DELETE', url, data);
    }
  }, {
    key: 'head',
    value: function head(url, data) {
      return this.do('HEAD', url, data);
    }
  }, {
    key: 'do',
    value: function _do(method, url, data) {
      url = this.baseUrl + (url.charAt(0) === '/' ? '' : location.pathname) + url;
      var that = this;
      return new Promise(function (resolve, reject) {
        var xhr = new XMLHttpRequest();
        xhr.open(method, url, true);
        xhr.setRequestHeader('Content-Type', 'application/json');
        for (var key in that.upHeaders) {
          xhr.setRequestHeader(key, that.upHeaders[key]);
        }
        xhr.onload = function () {
          that.xhr = xhr;
          if (this.status === 200) {
            var _iteratorNormalCompletion = true;
            var _didIteratorError = false;
            var _iteratorError = undefined;

            try {
              for (var _iterator = that.transmitHeaders[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                var _key = _step.value;

                var value = xhr.getResponseHeader(_key);
                if (value) {
                  that.upHeaders[_key] = value;
                }
              }
            } catch (err) {
              _didIteratorError = true;
              _iteratorError = err;
            } finally {
              try {
                if (!_iteratorNormalCompletion && _iterator.return) {
                  _iterator.return();
                }
              } finally {
                if (_didIteratorError) {
                  throw _iteratorError;
                }
              }
            }

            var _data = void 0;
            try {
              _data = JSON.parse(this.responseText);
            } catch (e) {
              _data = this.responseText;
            }
            resolve(_data, this);
          } else {
            resolve(null, this);
          }
        };
        xhr.onerror = function (err) {
          reject(err);
        };

        if (data) {
          try {
            data = JSON.stringify(data);
          } catch (e) {}
        }
        xhr.send(data);
      });
    }
  }]);
  return _class;
}();

function _setVarsValue(vars, value, datas) {
  var data = datas;
  for (var i = 0; i < vars.length - 1; i++) {
    data = data[vars[i]];
    if (!data) return;
  }
  data[vars[vars.length - 1]] = value;
}

// function _makeString (str, datas) {
//   let defines = ''
//   for (let key in datas) {
//     defines += 'let ' + key + ' = datas["' + key + '"];'
//   }
//   str = str.replace(/\{(.+?)\}/g, function (full, part1) {
//     try {
//       let result = eval(defines + 'eval(part1)')
//       if (result === undefined || result === null) return ''
//       return result
//     } catch (err) {
//       return full
//     }
//   })
//   return str
// }

function _makeString(str, datas) {
  var args = [];
  var values = [];
  for (var key in datas) {
    args.push(key);
    values.push(datas[key]);
  }
  var lastArgsIndex = args.length;
  args.push('return null');

  str = str.replace(/\{(.+?)\}/g, function (full, part1) {
    try {
      args[lastArgsIndex] = 'return ' + part1;
      var func = Function.constructor.apply(null, args);
      var result = func.apply(null, values);
      if (result === undefined || result === null) return '';
      return result;
    } catch (err) {
      return full;
    }
  });
  return str;
}

var _class$2 = function () {
  function _class() {
    classCallCheck(this, _class);
  }

  createClass(_class, [{
    key: 'refresh',
    value: function refresh(node, datas) {
      // 仅处理属性
      if (node.vars) {
        for (var key in node.vars) {
          var newValue = _makeString(node.vars[key], datas);
          if (key === 'className' || key === 'data') {
            node[key] = newValue;
          } else {
            node.setAttribute(key, newValue);
          }
        }
      }
      this.make(node, datas);
    }
  }, {
    key: 'make',
    value: function make(targetNode, datas) {
      var _this = this;

      var nodes = [];
      var _iteratorNormalCompletion = true;
      var _didIteratorError = false;
      var _iteratorError = undefined;

      try {
        for (var _iterator = targetNode.childNodes[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
          var node = _step.value;

          nodes.push(node);
        }
      } catch (err) {
        _didIteratorError = true;
        _iteratorError = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion && _iterator.return) {
            _iterator.return();
          }
        } finally {
          if (_didIteratorError) {
            throw _iteratorError;
          }
        }
      }

      var _iteratorNormalCompletion2 = true;
      var _didIteratorError2 = false;
      var _iteratorError2 = undefined;

      try {
        var _loop = function _loop() {
          var node = _step2.value;

          if (node.isTplMaked) {
            node.parentElement.removeChild(node);
            return 'continue';
          }

          // 初始化循环和条件
          if (node.attributes && (node.attributes.each || node.attributes['if'])) {
            var which = node.attributes.each ? 'each' : 'if';
            var cdom = document.createComment('');
            if (node.attributes.each) {
              cdom.eachs = { index: 'index', item: 'item', items: '' };
              var eachs = node.getAttribute('each').split(/\s*,\s*|\s+in\s+/i).reverse();
              cdom.eachs.items = eachs[0];
              if (eachs.length > 1) cdom.eachs.item = eachs[1];
              if (eachs.length > 2) cdom.eachs.index = eachs[2];
            } else {
              cdom['ifs'] = node.getAttribute('if');
            }
            node.removeAttribute(which);
            cdom.htmlTpl = node.outerHTML;
            cdom.parentTagName = node.parentElement.tagName;
            node.parentElement.insertBefore(cdom, node);
            node.parentElement.removeChild(node);
            node = cdom; // 让下面可以立刻处理
          }

          // 初始化属性
          if (node.vars === undefined) {
            node.vars = false;
            var vars = {};
            if (node.attributes) {
              // 待处理列表
              // for (let attr of node.attributes) {
              for (var i in node.attributes) {
                var attr = node.attributes[i];
                if (attr.value && attr.value.indexOf('{') !== -1 && attr.value.indexOf('}') !== -1) {
                  if (attr.name === 'class') {
                    vars['className'] = attr.value;
                    node.className = '';
                  } else {
                    vars[attr.name] = attr.value;
                    node.setAttribute(attr.name, '');
                  }
                  if (!node.vars) node.vars = vars;
                }
              }
            }

            // 处理 TextNode
            if (node.data && node.data.indexOf('{') !== -1 && node.data.indexOf('}') !== -1) {
              vars['data'] = node.data;
              node.data = '';
              if (!node.vars) node.vars = vars;
            }
          }

          // 处理属性
          if (node.vars) {
            for (var key in node.vars) {
              var newValue = _makeString(node.vars[key], datas);
              if (key === 'className' || key === 'data') {
                node[key] = newValue;
              } else {
                node.setAttribute(key, newValue);
              }
            }
          }

          // 初始化数据绑定 bind
          if (node.binds === undefined) {
            node.binds = false;
            if (node.attributes && node.attributes.bind) {
              node.binds = node.attributes.bind.value.split('.');
              switch (node.tagName) {
                case 'INPUT':
                case 'TEXTAREA':
                  node.addEventListener('change', function (e) {
                    var v = null;
                    if (node.type === 'checkbox') {
                      v = e.target.getAttribute('checked') === null;
                    } else if (node.type === 'radio') {
                      if (e.target.checked !== null) {
                        v = e.target.value;
                      }
                    } else {
                      v = e.target.value;
                    }
                    if (v !== null) {
                      _setVarsValue(e.target.binds, v, node.bindDatas);
                      var onbind = e.target.getAttribute('onbind');
                      if (onbind) {
                        var func = Function.constructor.apply(null, [onbind]);
                        func(e);
                      }
                    }
                  });
                  break;
              }
            }
          }

          // 处理属性
          if (node.binds) {
            node.bindDatas = {};
            for (var k in datas) {
              node.bindDatas[k] = datas[k];
            }
            switch (node.tagName) {
              case 'INPUT':
              case 'TEXTAREA':
                var data = datas;
                var _iteratorNormalCompletion3 = true;
                var _didIteratorError3 = false;
                var _iteratorError3 = undefined;

                try {
                  for (var _iterator3 = node.binds[Symbol.iterator](), _step3; !(_iteratorNormalCompletion3 = (_step3 = _iterator3.next()).done); _iteratorNormalCompletion3 = true) {
                    var _k = _step3.value;

                    data = data[_k];
                    if (!data) break;
                  }
                } catch (err) {
                  _didIteratorError3 = true;
                  _iteratorError3 = err;
                } finally {
                  try {
                    if (!_iteratorNormalCompletion3 && _iterator3.return) {
                      _iterator3.return();
                    }
                  } finally {
                    if (_didIteratorError3) {
                      throw _iteratorError3;
                    }
                  }
                }

                if (node.type === 'checkbox') {
                  if (data === true || data === 'true' || data === 1) {
                    node.setAttribute('checked', '');
                  } else {
                    node.removeAttribute('checked');
                  }
                }
                if (node.type === 'radio') {
                  if (data === node.value) {
                    node.checked = true;
                  } else {
                    node.checked = false;
                  }
                } else {
                  node.value = data || '';
                }
            }
          }

          // 不处理SUBVIEW
          if (node.id && node.id.startsWith('SUBVIEW_')) {
            return 'continue';
          }

          // 处理条件判断
          if (node['ifs']) {
            var isShow = _makeString('{' + node['ifs'] + '}', datas);
            if (isShow && isShow.charAt(0) !== '{' && isShow !== 'false' && isShow !== '0' && isShow !== 'undefined' && isShow !== 'null') {
              var dom = document.createElement(node.parentTagName || 'div');
              dom.innerHTML = node.htmlTpl;
              _this.make(dom, datas);
              dom.firstChild.isTplMaked = true;
              node.parentElement.insertBefore(dom.firstChild, node);
              // this.make(node.previousSibling, datas)
            }
            return 'continue';
          }

          // 处理循环
          if (node.eachs) {
            if (node.eachs.items) {
              // 根据 items 找到数据
              var itemsArr = node.eachs.items.split('.');
              var itemsData = datas;
              var _iteratorNormalCompletion4 = true;
              var _didIteratorError4 = false;
              var _iteratorError4 = undefined;

              try {
                for (var _iterator4 = itemsArr[Symbol.iterator](), _step4; !(_iteratorNormalCompletion4 = (_step4 = _iterator4.next()).done); _iteratorNormalCompletion4 = true) {
                  var itemsA = _step4.value;

                  itemsData = itemsData[itemsA];
                  if (itemsData === undefined) break;
                }
                // 处理数据
              } catch (err) {
                _didIteratorError4 = true;
                _iteratorError4 = err;
              } finally {
                try {
                  if (!_iteratorNormalCompletion4 && _iterator4.return) {
                    _iterator4.return();
                  }
                } finally {
                  if (_didIteratorError4) {
                    throw _iteratorError4;
                  }
                }
              }

              if (itemsData) {
                for (var index in itemsData) {
                  var _dom = document.createElement(node.parentTagName || 'div');
                  _dom.innerHTML = node.htmlTpl;
                  datas[node.eachs.index] = index;
                  datas[node.eachs.item] = itemsData[index];
                  _this.make(_dom, datas);
                  _dom.firstChild.isTplMaked = true;
                  node.parentElement.insertBefore(_dom.firstChild, node);
                  // this.make(node.previousSibling, datas)
                }
              }
            }
            return 'continue';
          }

          // 递归处理子集
          _this.make(node, datas);
        };

        for (var _iterator2 = nodes[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
          var _ret = _loop();

          if (_ret === 'continue') continue;
        }
      } catch (err) {
        _didIteratorError2 = true;
        _iteratorError2 = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion2 && _iterator2.return) {
            _iterator2.return();
          }
        } finally {
          if (_didIteratorError2) {
            throw _iteratorError2;
          }
        }
      }
    }
  }]);
  return _class;
}();

var _typeof$1 = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? function (obj) {
  return typeof obj;
} : function (obj) {
  return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj;
};











var classCallCheck$1 = function (instance, Constructor) {
  if (!(instance instanceof Constructor)) {
    throw new TypeError("Cannot call a class as a function");
  }
};

var createClass$1 = function () {
  function defineProperties(target, props) {
    for (var i = 0; i < props.length; i++) {
      var descriptor = props[i];
      descriptor.enumerable = descriptor.enumerable || false;
      descriptor.configurable = true;
      if ("value" in descriptor) descriptor.writable = true;
      Object.defineProperty(target, descriptor.key, descriptor);
    }
  }

  return function (Constructor, protoProps, staticProps) {
    if (protoProps) defineProperties(Constructor.prototype, protoProps);
    if (staticProps) defineProperties(Constructor, staticProps);
    return Constructor;
  };
}();

var Route = function () {
  function Route() {
    classCallCheck$1(this, Route);

    this.routeUrl = '';
    this._routeHistoryPos = -1;
    this._routeHistories = [];
    this._binds = {};
    this._spaceChars = ['/', '#', '!', '`', '~', '@', '%', '^', '*', ';', '\\', ' '];
  }

  createClass$1(Route, [{
    key: 'go',
    value: function go(url, args) {
      var paths = [];
      var urls = [];
      var currentPaths = this._routeHistoryPos >= 0 ? this._routeHistories[this._routeHistoryPos] : [];

      if (url === 'up') {
        // 向上一级
        if (currentPaths.length > 0) urls = currentPaths.slice(0, currentPaths.length - 1);
      } else if (typeof url === 'string') {
        // url路由
        var space = url.charAt(0);
        if (space === '.') space = url.charAt(1);
        if (space === '.') space = url.charAt(2);
        if (this._spaceChars.indexOf(space) === -1) {
          // 追加模式
          paths = JSON.parse(JSON.stringify(currentPaths));
          urls = [url];
        } else if (url.charAt(0) === '.') {
          // 追加模式
          paths = JSON.parse(JSON.stringify(currentPaths));
          urls = url.split(space);
        } else {
          urls = url.split(space);
        }
      } else if (typeof url === 'number') {
        // 历史路由
        this._routeHistoryPos += url;
        if (this._routeHistoryPos < 0) this._routeHistoryPos = 0;
        if (this._routeHistoryPos > this._routeHistories.length - 1) this._routeHistoryPos = this._routeHistories.length - 1;
        if (this._routeHistoryPos >= 0) urls = this._routeHistories[this._routeHistoryPos];
      } else if (url instanceof Array) {
        // 数组路由
        urls = url;
      }

      var tmpNames = [];
      var _iteratorNormalCompletion = true;
      var _didIteratorError = false;
      var _iteratorError = undefined;

      try {
        for (var _iterator = paths[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
          var tmpPath = _step.value;

          tmpNames.push(tmpPath.name);
        }
      } catch (err) {
        _didIteratorError = true;
        _iteratorError = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion && _iterator.return) {
            _iterator.return();
          }
        } finally {
          if (_didIteratorError) {
            throw _iteratorError;
          }
        }
      }

      var path = null;
      var _iteratorNormalCompletion2 = true;
      var _didIteratorError2 = false;
      var _iteratorError2 = undefined;

      try {
        for (var _iterator2 = urls[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
          var _url = _step2.value;

          if (!_url) continue;
          if (typeof _url === 'string') {
            if (_url === '.') continue;
            if (_url === '..') {
              if (paths.length > 0) {
                paths = paths.slice(0, paths.length - 1);
                // names = names.slice(0, names.length - 1)
                // newUrls = newUrls.slice(0, newUrls.length - 1)
              }
              continue;
            }

            _url = decodeURIComponent(_url);
            path = { args: {} };
            var argsPos = void 0;
            if ((argsPos = _url.indexOf('?')) !== -1) {
              path.name = _url.substr(0, argsPos);
              var argsA = _url.substr(argsPos + 1).split('&');
              var _iteratorNormalCompletion5 = true;
              var _didIteratorError5 = false;
              var _iteratorError5 = undefined;

              try {
                for (var _iterator5 = argsA[Symbol.iterator](), _step5; !(_iteratorNormalCompletion5 = (_step5 = _iterator5.next()).done); _iteratorNormalCompletion5 = true) {
                  var argA = _step5.value;

                  if (!argA) continue;
                  var argPos = argA.indexOf('=');
                  if (argPos !== -1) {
                    path.args[argA.substr(0, argPos)] = argA.substr(argPos + 1);
                  }
                }
              } catch (err) {
                _didIteratorError5 = true;
                _iteratorError5 = err;
              } finally {
                try {
                  if (!_iteratorNormalCompletion5 && _iterator5.return) {
                    _iterator5.return();
                  }
                } finally {
                  if (_didIteratorError5) {
                    throw _iteratorError5;
                  }
                }
              }
            } else if ((argsPos = _url.indexOf('{')) !== -1) {
              path.name = _url.substr(0, argsPos);
              try {
                path.args = JSON.parse(_url.substr(argsPos));
              } catch (err) {}
            } else {
              path.name = _url;
            }

            // 相同路径的旧节点如果有参数，继承过来
            tmpNames.push(path.name);
            var tmpPathName = tmpNames.join('.');
            // console.log([tmpPathName, paths.length, currentPaths[paths.length]])
            var prevSamePath = currentPaths[paths.length];
            if (prevSamePath && tmpPathName === prevSamePath.pathName && prevSamePath.args) {
              for (var _k in prevSamePath.args) {
                if (!path.args[_k]) path.args[_k] = prevSamePath.args[_k];
              }
            }
          } else {
            path = _url;
          }

          paths.push(path);
        }

        // 附加参数
      } catch (err) {
        _didIteratorError2 = true;
        _iteratorError2 = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion2 && _iterator2.return) {
            _iterator2.return();
          }
        } finally {
          if (_didIteratorError2) {
            throw _iteratorError2;
          }
        }
      }

      if (path !== null && args && (typeof args === 'undefined' ? 'undefined' : _typeof$1(args)) === 'object') {
        for (var k in args) {
          path.args[k] = args[k];
        }
      }

      // console.log(this.routeUrl);
      if (typeof url !== 'number') {
        if (this._routeHistoryPos < this._routeHistories.length - 1) {
          this._routeHistories = this._routeHistories.slice(0, this._routeHistoryPos + 1);
        }
        this._routeHistories.push(paths);
        if (this._routeHistories.length > 10) {
          this._routeHistories = this._routeHistories.slice(this._routeHistories.length - 11);
        }
        this._routeHistoryPos = this._routeHistories.length - 1;
      }

      var oldUrl = this.routeUrl;
      this.remakeRouteUrl();
      if (this.routeUrl === oldUrl) {
        // 相同的路由不触发事件
        return;
      }

      // // 重新计算 pathName、url
      // let newUrls = []
      // let names = []
      // for (let path of paths) {
      //   names.push(path.name)
      //   path.pathName = names.join('.')
      //
      //   path.url = path.name + (Object.getOwnPropertyNames(path.args).length ? JSON.stringify(path.args) : '')
      //   newUrls.push(path.url)
      // }
      // paths.last = path || {name: '', args: '', pathName: '', url: ''}
      //
      // // 寻找一个没出现过的字符作为间隔符
      // let space = '/'
      // let tmpNewUrlString = newUrls.join('')
      // for (let spaceChar of this._spaceChars) {
      //   if (tmpNewUrlString.indexOf(spaceChar) === -1) {
      //     space = spaceChar
      //     break
      //   }
      // }
      // let newUrl = space + newUrls.join(space)
      // if (newUrl === this.routeUrl) return
      // this.routeUrl = newUrl

      var that = this;
      return new Promise(function (resolve, reject) {
        var pendingTargets = []; // 需要回调的对象
        for (var bindKey in that._binds) {
          if (bindKey === '*') {
            // all match
            pendingTargets.push(that._binds[bindKey]);
          } else {
            var partInfo = that._binds[bindKey].partInfo;
            if (!paths.last && paths.length > 0) paths.last = paths[paths.length - 1];
            if (paths.last) {
              if (partInfo.pos === -1) {
                if (bindKey === paths.last.pathName) {
                  pendingTargets.push(that._binds[bindKey]);
                }
              } else {
                if (partInfo.k1 && !partInfo.k2 && paths.last.pathName.startsWith(partInfo.k1) || partInfo.k2 && !partInfo.k1 && paths.last.pathName.endsWith(partInfo.k2) || partInfo.k1 && partInfo.k2 && paths.last.pathName.startsWith(partInfo.k1) && paths.last.pathName.endsWith(partInfo.k2)) {
                  pendingTargets.push(that._binds[bindKey]);
                }
              }
            }
          }
        }

        var _iteratorNormalCompletion3 = true;
        var _didIteratorError3 = false;
        var _iteratorError3 = undefined;

        try {
          for (var _iterator3 = pendingTargets[Symbol.iterator](), _step3; !(_iteratorNormalCompletion3 = (_step3 = _iterator3.next()).done); _iteratorNormalCompletion3 = true) {
            var targets = _step3.value;
            var _iteratorNormalCompletion4 = true;
            var _didIteratorError4 = false;
            var _iteratorError4 = undefined;

            try {
              for (var _iterator4 = targets[Symbol.iterator](), _step4; !(_iteratorNormalCompletion4 = (_step4 = _iterator4.next()).done); _iteratorNormalCompletion4 = true) {
                var target = _step4.value;

                if (target.func !== null && typeof target.func === 'function') {
                  target.func.call(target.object, paths);
                }
              }
            } catch (err) {
              _didIteratorError4 = true;
              _iteratorError4 = err;
            } finally {
              try {
                if (!_iteratorNormalCompletion4 && _iterator4.return) {
                  _iterator4.return();
                }
              } finally {
                if (_didIteratorError4) {
                  throw _iteratorError4;
                }
              }
            }
          }
        } catch (err) {
          _didIteratorError3 = true;
          _iteratorError3 = err;
        } finally {
          try {
            if (!_iteratorNormalCompletion3 && _iterator3.return) {
              _iterator3.return();
            }
          } finally {
            if (_didIteratorError3) {
              throw _iteratorError3;
            }
          }
        }

        resolve(paths);
      });
    }

    // 生成URL

  }, {
    key: 'remakeRouteUrl',
    value: function remakeRouteUrl() {
      var paths = this._routeHistories[this._routeHistoryPos];
      var newUrls = [];
      var names = [];
      var _iteratorNormalCompletion6 = true;
      var _didIteratorError6 = false;
      var _iteratorError6 = undefined;

      try {
        for (var _iterator6 = paths[Symbol.iterator](), _step6; !(_iteratorNormalCompletion6 = (_step6 = _iterator6.next()).done); _iteratorNormalCompletion6 = true) {
          var path = _step6.value;

          names.push(path.name);
          path.pathName = names.join('.');

          path.url = path.name + (Object.getOwnPropertyNames(path.args).length ? JSON.stringify(path.args) : '');
          newUrls.push(path.url);
          paths.last = path;
        }
      } catch (err) {
        _didIteratorError6 = true;
        _iteratorError6 = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion6 && _iterator6.return) {
            _iterator6.return();
          }
        } finally {
          if (_didIteratorError6) {
            throw _iteratorError6;
          }
        }
      }

      if (!paths.last) paths.last = { name: '', args: '', pathName: '', url: ''

        // 寻找一个没出现过的字符作为间隔符
      };var space = '/';
      var tmpNewUrlString = newUrls.join('');
      var _iteratorNormalCompletion7 = true;
      var _didIteratorError7 = false;
      var _iteratorError7 = undefined;

      try {
        for (var _iterator7 = this._spaceChars[Symbol.iterator](), _step7; !(_iteratorNormalCompletion7 = (_step7 = _iterator7.next()).done); _iteratorNormalCompletion7 = true) {
          var spaceChar = _step7.value;

          if (tmpNewUrlString.indexOf(spaceChar) === -1) {
            space = spaceChar;
            break;
          }
        }
      } catch (err) {
        _didIteratorError7 = true;
        _iteratorError7 = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion7 && _iterator7.return) {
            _iterator7.return();
          }
        } finally {
          if (_didIteratorError7) {
            throw _iteratorError7;
          }
        }
      }

      var newUrl = space + newUrls.join(space);
      this.routeUrl = newUrl;
    }

    // 绑定数据变化通知

  }, {
    key: 'bind',
    value: function bind(keyOrKeys, target) {
      if (!(keyOrKeys instanceof Array)) keyOrKeys = [keyOrKeys];
      var bindTarget = {
        object: null,
        func: null,
        keys: keyOrKeys

        // 构建回调对象
      };if (typeof target === 'function') {
        // direct call function
        bindTarget.func = target;
      } else if ((typeof target === 'undefined' ? 'undefined' : _typeof$1(target)) === 'object') {
        // call object's [set or setData]
        if (target['onRoute'] && typeof target['onRoute'] === 'function') {
          bindTarget.object = target;
          bindTarget.func = target['onRoute'];
        }
      }

      var _iteratorNormalCompletion8 = true;
      var _didIteratorError8 = false;
      var _iteratorError8 = undefined;

      try {
        for (var _iterator8 = keyOrKeys[Symbol.iterator](), _step8; !(_iteratorNormalCompletion8 = (_step8 = _iterator8.next()).done); _iteratorNormalCompletion8 = true) {
          var key = _step8.value;

          if (!this._binds[key]) this._binds[key] = [];
          this._binds[key].push(bindTarget);

          var partInfo = this._binds[key].partInfo;
          if (partInfo === undefined) {
            partInfo = {};
            partInfo.pos = key.indexOf('*');
            if (partInfo.pos !== -1) {
              partInfo.k1 = key.substr(0, partInfo.pos);
              partInfo.k2 = key.substr(partInfo.pos + 1);
            }
            this._binds[key].partInfo = partInfo;
          }
        }
      } catch (err) {
        _didIteratorError8 = true;
        _iteratorError8 = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion8 && _iterator8.return) {
            _iterator8.return();
          }
        } finally {
          if (_didIteratorError8) {
            throw _iteratorError8;
          }
        }
      }
    }

    // 取消绑定

  }, {
    key: 'unbind',
    value: function unbind(keyOrKeys, target) {
      if (!(keyOrKeys instanceof Array)) {
        keyOrKeys = [keyOrKeys];
      }
      var jsonKeys = JSON.stringify(keyOrKeys);
      var _iteratorNormalCompletion9 = true;
      var _didIteratorError9 = false;
      var _iteratorError9 = undefined;

      try {
        for (var _iterator9 = keyOrKeys[Symbol.iterator](), _step9; !(_iteratorNormalCompletion9 = (_step9 = _iterator9.next()).done); _iteratorNormalCompletion9 = true) {
          var key = _step9.value;

          var bindTargets = this._binds[key];
          if (bindTargets) {
            for (var i = bindTargets.length - 1; i >= 0; i--) {
              var bindTarget = bindTargets[i];
              if ((bindTarget.object === target || bindTarget.func === target) && JSON.stringify(bindTarget.keys) === jsonKeys) {
                this._binds[key].splice(i, 1);
              }
            }
          }
        }
      } catch (err) {
        _didIteratorError9 = true;
        _iteratorError9 = err;
      } finally {
        try {
          if (!_iteratorNormalCompletion9 && _iterator9.return) {
            _iterator9.return();
          }
        } finally {
          if (_didIteratorError9) {
            throw _iteratorError9;
          }
        }
      }
    }
  }]);
  return Route;
}();

/*
global location
global XMLHttpRequest
 */

var tpl = new _class$2();

function _$(selector) {
  return document.querySelector('#SUBVIEW_' + this.parentPathName + ' ' + (selector || ''));
}

function _moveNodes(to, from) {
  while (from.childNodes.length > 0) {
    to.appendChild(from.childNodes[0]);
  }
}

function _setData(values) {
  var routeChanged = false;
  var path = null;
  for (var key in values) {
    this.data[key] = values[key];
    // 绑定的参数反向同步到路由
    if (this.routeBinds && this.routeBinds.indexOf(key) !== -1) {
      // 查找当前路由节点
      if (path === null) {
        var _iteratorNormalCompletion = true;
        var _didIteratorError = false;
        var _iteratorError = undefined;

        try {
          for (var _iterator = this.route._routeHistories[this.route._routeHistoryPos][Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
            var p = _step.value;

            if (p.pathName === this.pathName) {
              path = p;
            }
          }
        } catch (err) {
          _didIteratorError = true;
          _iteratorError = err;
        } finally {
          try {
            if (!_iteratorNormalCompletion && _iterator.return) {
              _iterator.return();
            }
          } finally {
            if (_didIteratorError) {
              throw _iteratorError;
            }
          }
        }
      }
      // 比较数据是否变更
      if (path) {
        if (path.args[key] !== this.data[key]) {
          path.args[key] = this.data[key];
          routeChanged = true;
        }
      }
    }

    // 数据变化后反向同步到路由
    if (routeChanged) {
      this.route.remakeRouteUrl();
      location.hash = '#' + this.route.routeUrl;
    }
  }
  return this.refreshView();
}

function _refreshView() {
  var _this = this;

  return new Promise(function (resolve, reject) {
    if (!_this._refreshViewCallbacks) _this._refreshViewCallbacks = [];
    _this._refreshViewCallbacks.push([resolve, reject]);
    if (!_this._refreshViewTID) {
      setTimeout(function () {
        try {
          var datas = this.datas || {};
          datas.data = this.data;
          tpl.make(this.$(), datas);
          var _iteratorNormalCompletion2 = true;
          var _didIteratorError2 = false;
          var _iteratorError2 = undefined;

          try {
            for (var _iterator2 = this._refreshViewCallbacks[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
              var callbacks = _step2.value;

              callbacks[0]();
            }
          } catch (err) {
            _didIteratorError2 = true;
            _iteratorError2 = err;
          } finally {
            try {
              if (!_iteratorNormalCompletion2 && _iterator2.return) {
                _iterator2.return();
              }
            } finally {
              if (_didIteratorError2) {
                throw _iteratorError2;
              }
            }
          }
        } catch (err) {
          var _iteratorNormalCompletion3 = true;
          var _didIteratorError3 = false;
          var _iteratorError3 = undefined;

          try {
            for (var _iterator3 = this._refreshViewCallbacks[Symbol.iterator](), _step3; !(_iteratorNormalCompletion3 = (_step3 = _iterator3.next()).done); _iteratorNormalCompletion3 = true) {
              var _callbacks = _step3.value;

              _callbacks[1](err);
            }
          } catch (err) {
            _didIteratorError3 = true;
            _iteratorError3 = err;
          } finally {
            try {
              if (!_iteratorNormalCompletion3 && _iterator3.return) {
                _iterator3.return();
              }
            } finally {
              if (_didIteratorError3) {
                throw _iteratorError3;
              }
            }
          }
        } finally {
          this._refreshViewTID = 0;
          this._refreshViewCallbacks = [];
        }
      }.bind(_this), 0);
    }
  });
}

var _class$3 = function (_Route) {
  inherits(_class, _Route);

  function _class(states) {
    classCallCheck(this, _class);

    var _this2 = possibleConstructorReturn(this, (_class.__proto__ || Object.getPrototypeOf(_class)).call(this));

    _this2.states = states;
    _this2._prevPaths = [];
    return _this2;
  }

  createClass(_class, [{
    key: 'bindHash',
    value: function bindHash() {
      var _this3 = this;

      var that = this;
      this.bind('*', function (paths) {
        if (that.makeRoute(paths)) {
          location.hash = '#' + _this3.routeUrl;
        } else {
          setTimeout(function () {
            that.go(-1);
          });
        }
      });
      window.addEventListener('hashchange', function () {
        _this3.go(location.hash.substr(1));
      });
    }
  }, {
    key: 'makeRoute',
    value: function makeRoute(paths) {
      // 预处理
      var parentView = this.Root;
      var availablePaths = [];
      var samePos = -1;
      if (window._cachedViews === undefined) window._cachedViews = { ROOT: this.Root };
      for (var i = 0; i < paths.length; i++) {
        var path = paths[i];
        var prevPath = this._prevPaths[i];
        var view = window._cachedViews[path.pathName];
        if (!view) {
          if (parentView.getSubView) view = parentView.getSubView(path.name);
          if (!view) break;
          // 生成dom节点
          view.dom = document.createElement('div');
          if (view.html) {
            if (/\.html\s*$/.test(view.html)) {
              var xhr = new XMLHttpRequest();
              xhr.open('GET', view.html, false);
              xhr.send();
              view.html = xhr.responseText;
            }
            view.dom.innerHTML = view.html.replace(/\$this/g, 'window._cachedViews[\'' + path.pathName + '\']').replace('id="SUBVIEW"', 'id="SUBVIEW_' + path.pathName + '"');
          }
          window._cachedViews[path.pathName] = view;
          // 新路由节点调用 onCreate
          if (view.onCreate) {
            view.onCreate(path);
          }
          if (!view.data) {
            view.data = {};
          }
          view.route = this;
          view.$ = _$;
          view.setData = _setData;
          view.refreshView = _refreshView;
        }
        view.pathName = path.pathName;
        availablePaths.push(path);
        parentView = view;
        if (prevPath && prevPath.url === path.url && samePos === i - 1) {
          samePos = i;
        }
      }

      // 旧路由中不一样的部分调用 canHide，确认允许跳转
      for (var _i = this._prevPaths.length - 1; _i > samePos; _i--) {
        var _path = _i === -1 ? { pathName: 'ROOT' } : this._prevPaths[_i];
        var _view = _i === -1 ? this.Root : window._cachedViews[_path.pathName];
        if (_view.canHide) {
          if (!_view.canHide(_path, paths)) {
            return false;
          }
        }
      }

      // 旧路由中不一样的部分调用 onHide
      for (var _i2 = this._prevPaths.length - 1; _i2 > samePos; _i2--) {
        var _path2 = _i2 === -1 ? { pathName: 'ROOT' } : this._prevPaths[_i2];
        var _view2 = _i2 === -1 ? this.Root : window._cachedViews[_path2.pathName];
        var _prevPath = (_i2 === 0 ? { pathName: 'ROOT' } : this._prevPaths[_i2 - 1]) || null;
        // let prevView = prevPath ? _cachedViews[prevPath.pathName] : null
        if (_view2.stateRegisters) {
          for (var bindKey in _view2.stateRegisters) {
            this.states.unbind(bindKey, _view2.stateRegisters[bindKey]);
          }
        }
        if (_view2.stateBinds) {
          this.states.unbind(_view2.stateBinds, _view2);
        }

        if (_view2.onHide) _view2.onHide(_path2);
        if (_prevPath) {
          var container = document.querySelector('#SUBVIEW_' + _prevPath.pathName);
          if (container) {
            _moveNodes(_view2.dom, container);
          }
        }
      }

      // 新路由中不一样的部分调用 onShow
      for (var _i3 = samePos; _i3 < availablePaths.length; _i3++) {
        var _path3 = _i3 === -1 ? { pathName: 'ROOT' } : availablePaths[_i3];
        var _view3 = window._cachedViews[_path3.pathName];
        var nextPath = availablePaths[_i3 + 1] || null;
        var nextView = nextPath ? window._cachedViews[nextPath.pathName] : null;
        if (nextView) {
          nextView.parentPathName = _path3.pathName;
          if (_view3.data) {
            // 默认维护 subName
            _view3.data.subName = nextPath.name;
          }
          // nextView.$ = this.$
          var _container = document.querySelector('#SUBVIEW_' + _path3.pathName);
          if (_container) {
            _moveNodes(_container, nextView.dom);
          }
        }

        // 路由参数并入 view.data
        if (_view3.routeBinds) {
          var _iteratorNormalCompletion4 = true;
          var _didIteratorError4 = false;
          var _iteratorError4 = undefined;

          try {
            for (var _iterator4 = _view3.routeBinds[Symbol.iterator](), _step4; !(_iteratorNormalCompletion4 = (_step4 = _iterator4.next()).done); _iteratorNormalCompletion4 = true) {
              var k = _step4.value;

              _view3.data[k] = _path3.args[k];
            }
          } catch (err) {
            _didIteratorError4 = true;
            _iteratorError4 = err;
          } finally {
            try {
              if (!_iteratorNormalCompletion4 && _iterator4.return) {
                _iterator4.return();
              }
            } finally {
              if (_didIteratorError4) {
                throw _iteratorError4;
              }
            }
          }
        }

        if (_i3 > samePos) {
          if (_view3.onShow) _view3.onShow(_path3, nextPath, nextView);
          // 自动绑定state，并触发一次所有绑定事件
          if (_view3.stateRegisters) {
            for (var _bindKey in _view3.stateRegisters) {
              var bindTarget = _view3.stateRegisters[_bindKey];
              this.states.bind(_bindKey, bindTarget);
              if (typeof bindTarget === 'function') {
                bindTarget(this.states.get(_bindKey));
              } else if (bindTarget instanceof Array && bindTarget.length === 2) {
                if (typeof bindTarget[1] === 'function') {
                  bindTarget[1].apply(bindTarget[0], this.states.get(_bindKey));
                } else if (typeof bindTarget[0][bindTarget[1]] === 'function') {
                  bindTarget[0][bindTarget[1]].apply(bindTarget[0], this.states.get(_bindKey));
                }
              } else if ((typeof bindTarget === 'undefined' ? 'undefined' : _typeof(bindTarget)) === 'object' && typeof bindTarget.setData === 'function') {
                bindTarget.setData(this.states.get(_bindKey));
              }
            }
          }
          if (_view3.stateBinds) {
            this.states.bind(_view3.stateBinds, _view3);
            var _iteratorNormalCompletion5 = true;
            var _didIteratorError5 = false;
            var _iteratorError5 = undefined;

            try {
              for (var _iterator5 = _view3.stateBinds[Symbol.iterator](), _step5; !(_iteratorNormalCompletion5 = (_step5 = _iterator5.next()).done); _iteratorNormalCompletion5 = true) {
                var bind = _step5.value;

                if (typeof bind === 'string') {
                  _view3.data[bind] = this.states.state[bind];
                } else if (bind instanceof Array && bind.length === 2) {
                  _view3.data[bind[1]] = this.states.state[bind[0]];
                }
              }
            } catch (err) {
              _didIteratorError5 = true;
              _iteratorError5 = err;
            } finally {
              try {
                if (!_iteratorNormalCompletion5 && _iterator5.return) {
                  _iterator5.return();
                }
              } finally {
                if (_didIteratorError5) {
                  throw _iteratorError5;
                }
              }
            }
          }
        }

        if (_view3.refreshView) _view3.refreshView();
      }

      this._prevPaths = availablePaths;

      return true;
    }
  }]);
  return _class;
}(Route);

exports.Http = _class;
exports.Tpl = _class$2;
exports.Route = _class$3;

Object.defineProperty(exports, '__esModule', { value: true });

})));
