(function (global, factory) {
	typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports) :
	typeof define === 'function' && define.amd ? define(['exports'], factory) :
	(factory((global.svcState = global.svcState || {})));
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





var defineProperty = function (obj, key, value) {
  if (key in obj) {
    Object.defineProperty(obj, key, {
      value: value,
      enumerable: true,
      configurable: true,
      writable: true
    });
  } else {
    obj[key] = value;
  }

  return obj;
};

var State = function () {
  function State(whichDataResolved) {
    classCallCheck(this, State);

    this.state = {};
    this._whichDataResolved = whichDataResolved || 'changes'; // changes binds all none
    this._changedStates = {};
    this._setResolves = [];
    this._binds = {};
    this._partBindsCache = {};
    this._tid = 0;
  }

  createClass(State, [{
    key: '_testIfMatch',
    value: function _testIfMatch(partKey, realKey) {
      var pos = partKey.indexOf('*');
      if (pos !== -1) {
        var k1 = partKey.substr(0, pos);
        var k2 = partKey.substr(pos + 1);
        return k1 && !k2 && realKey.startsWith(k1) || k2 && !k1 && realKey.endsWith(k2) || k1 && k2 && realKey.startsWith(k1) && realKey.endsWith(k2);
      }
    }
  }, {
    key: '_notice',
    value: function _notice() {
      var pendingTargets = [];
      for (var bindKey in this._binds) {
        if (bindKey === '*') {
          // all match
          pendingTargets.push(this._binds[bindKey]);
        } else {
          var pos = bindKey.indexOf('*');
          if (pos === -1) {
            if (this._changedStates[bindKey] !== undefined && pendingTargets.indexOf(this._binds[bindKey]) === -1) {
              // exact match
              pendingTargets.push(this._binds[bindKey]);
            }
          } else {
            var k1 = bindKey.substr(0, pos);
            var k2 = bindKey.substr(pos + 1);
            var partBinds = this._partBindsCache[bindKey];
            if (!partBinds) {
              partBinds = [];
              // make part binds cache
              for (var stateKey in this.state) {
                if (k1 && !k2 && stateKey.startsWith(k1) || k2 && !k1 && stateKey.endsWith(k2) || k1 && k2 && stateKey.startsWith(k1) && stateKey.endsWith(k2)) {
                  partBinds.push(stateKey);
                }
              }
              this._partBindsCache[bindKey] = partBinds;
            }

            var _iteratorNormalCompletion = true;
            var _didIteratorError = false;
            var _iteratorError = undefined;

            try {
              for (var _iterator = partBinds[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                var _stateKey = _step.value;

                if (this._changedStates[_stateKey] !== undefined && pendingTargets.indexOf(this._binds[bindKey]) === -1) {
                  // part match
                  pendingTargets.push(this._binds[bindKey]);
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
        }
      }

      var _iteratorNormalCompletion2 = true;
      var _didIteratorError2 = false;
      var _iteratorError2 = undefined;

      try {
        for (var _iterator2 = pendingTargets[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
          var targets = _step2.value;
          var _iteratorNormalCompletion4 = true;
          var _didIteratorError4 = false;
          var _iteratorError4 = undefined;

          try {
            for (var _iterator4 = targets[Symbol.iterator](), _step4; !(_iteratorNormalCompletion4 = (_step4 = _iterator4.next()).done); _iteratorNormalCompletion4 = true) {
              var target = _step4.value;

              var data = {};
              switch (this._whichDataResolved) {
                case 'changes':
                  data = this._changedStates;
                  break;
                case 'binds':
                  data = this.get(target.keys);
                  break;
                case 'all':
                  data = this.state;
                  break;
              }

              if (target.func === null && _typeof(target.object) === 'object') {
                for (var key in data) {
                  target.object[key] = data[key];
                }
              } else if (typeof target.func === 'function') {
                target.func.call(target.object, data);
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

      var _iteratorNormalCompletion3 = true;
      var _didIteratorError3 = false;
      var _iteratorError3 = undefined;

      try {
        for (var _iterator3 = this._setResolves[Symbol.iterator](), _step3; !(_iteratorNormalCompletion3 = (_step3 = _iterator3.next()).done); _iteratorNormalCompletion3 = true) {
          var resolve = _step3.value;

          resolve();
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

      this._changedStates = {};
      this._setResolves = [];
      this._tid = 0;
    }

    // 获取数据，传入数组一次取多个，也可以直接访问State对象不通过 get

  }, {
    key: 'get',
    value: function get$$1(keyOrKeys) {
      if (!(keyOrKeys instanceof Array)) {
        keyOrKeys = [keyOrKeys];
      }
      var result = {};
      var _iteratorNormalCompletion5 = true;
      var _didIteratorError5 = false;
      var _iteratorError5 = undefined;

      try {
        for (var _iterator5 = keyOrKeys[Symbol.iterator](), _step5; !(_iteratorNormalCompletion5 = (_step5 = _iterator5.next()).done); _iteratorNormalCompletion5 = true) {
          var key = _step5.value;

          if (key === '*') {
            return this.state;
          }
          if (key.indexOf('*') !== -1) {
            for (var stateKey in this.state) {
              if (this._testIfMatch(key, stateKey)) {
                result[stateKey] = this.state[stateKey];
              }
            }
          } else {
            result[key] = this.state[key];
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

      return result;
    }

    // 设置数据，传入对象一次存多个

  }, {
    key: 'set',
    value: function set$$1(keyOrValues, value) {
      var values = typeof keyOrValues === 'string' ? defineProperty({}, keyOrValues, value) : keyOrValues;
      var hasNew = false;
      for (var key in values) {
        if (this.state[key] === undefined) hasNew = true;
        this.state[key] = values[key];
        this._changedStates[key] = values[key];
      }

      if (hasNew) {
        // clear part binds when add new state
        this._partBindsCache = {};
      }

      if (!this._tid) {
        this._tid = setTimeout(this._notice.bind(this), 0);
      }

      var that = this;
      return new Promise(function (resolve, reject) {
        that._setResolves.push(resolve);
      });
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
      } else if (target instanceof Array && target.length === 2 && typeof target[1] === 'function') {
        bindTarget.object = target[0];
        bindTarget.func = target[1];
      } else if (target instanceof Array && target.length === 2 && typeof target[0][target[1]] === 'function') {
        bindTarget.object = target[0];
        bindTarget.func = target[0][target[1]];
      } else if ((typeof target === 'undefined' ? 'undefined' : _typeof(target)) === 'object') {
        // call object's [set or setData]
        var settingFunction = target['setState'] ? target['setState'] : target['setData'] ? target['setData'] : null;
        if (settingFunction && typeof settingFunction === 'function') {
          bindTarget.object = target;
          bindTarget.func = settingFunction;
        } else {
          bindTarget.object = target;
        }
      }

      var _iteratorNormalCompletion6 = true;
      var _didIteratorError6 = false;
      var _iteratorError6 = undefined;

      try {
        for (var _iterator6 = keyOrKeys[Symbol.iterator](), _step6; !(_iteratorNormalCompletion6 = (_step6 = _iterator6.next()).done); _iteratorNormalCompletion6 = true) {
          var key = _step6.value;

          if (!this._binds[key]) this._binds[key] = [];
          this._binds[key].push(bindTarget);
          this._partBindsCache[key] = null;
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
    }

    // 取消绑定

  }, {
    key: 'unbind',
    value: function unbind(keyOrKeys, target) {
      if (!(keyOrKeys instanceof Array)) {
        keyOrKeys = [keyOrKeys];
      }
      var jsonKeys = JSON.stringify(keyOrKeys);
      var _iteratorNormalCompletion7 = true;
      var _didIteratorError7 = false;
      var _iteratorError7 = undefined;

      try {
        for (var _iterator7 = keyOrKeys[Symbol.iterator](), _step7; !(_iteratorNormalCompletion7 = (_step7 = _iterator7.next()).done); _iteratorNormalCompletion7 = true) {
          var key = _step7.value;

          var bindTargets = this._binds[key];
          if (bindTargets) {
            for (var i = bindTargets.length - 1; i >= 0; i--) {
              var bindTarget = bindTargets[i];
              if ((bindTarget.object === target || bindTarget.func === target) && JSON.stringify(bindTarget.keys) === jsonKeys) {
                this._binds[key].splice(i, 1);
              }
            }
          }
          this._partBindsCache[key] = null;
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

      this._partBindsCache = {};
    }
  }]);
  return State;
}();

exports.State = State;

Object.defineProperty(exports, '__esModule', { value: true });

})));
