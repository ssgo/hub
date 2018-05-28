(function (global, factory) {
	typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports) :
	typeof define === 'function' && define.amd ? define(['exports'], factory) :
	(factory((global.svcAction = global.svcAction || {})));
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

var Action = function () {
  function Action(contexts) {
    classCallCheck(this, Action);

    this._targets = {};
    this._initedTargets = {};
    this._filters = [];
    this._contexts = [];
    this._lastTargetName = '';
    if (contexts instanceof Array) {
      this._contexts = contexts;
    } else if ((typeof contexts === 'undefined' ? 'undefined' : _typeof(contexts)) === 'object') {
      this._contexts = [contexts];
    }
  }

  createClass(Action, [{
    key: 'register',
    value: function register(targetName, targetObject) {
      this._targets[targetName] = targetObject;
    }
  }, {
    key: 'unregister',
    value: function unregister(targetName) {
      delete this._targets[targetName];
    }
  }, {
    key: 'addFilter',
    value: function addFilter(filterObject) {
      this._filters.push(filterObject);
    }
  }, {
    key: 'call',
    value: function call(callName, actionArgs) {
      var that = this;
      return new Promise(function (resolve, reject) {
        if (!actionArgs) actionArgs = {};
        var lastP = callName.lastIndexOf('.');
        if (lastP < 0) return reject(new Error('Action ' + callName + ' is not exists'));
        var targetName = callName.substr(0, lastP);
        var actionName = callName.substr(lastP + 1);
        if (!targetName && that._lastTargetName) {
          targetName = that._lastTargetName;
        }
        var targetObject = that._targets[targetName];
        if (actionName.charAt(0) === '_') return reject(new Error('Action ' + actionName + ' on ' + targetName + ' is private'));
        if (!targetObject) return reject(new Error('Target ' + targetName + ' is not exists when called action ' + actionName));
        if (!targetObject[actionName]) return reject(new Error('Action ' + actionName + ' on ' + targetName + ' is not exists'));
        that._lastTargetName = targetName;

        var actionContextArgs = that._makeContextArgs(targetObject, actionName, resolve, reject);
        if (targetObject['_init']) {
          if (!that._initedTargets[targetName]) {
            var initResolve = function initResolve() {
              that._initedTargets[targetName] = true;
              that._callAction(targetObject, actionName, actionContextArgs, actionArgs);
            };
            var initContextArgs = that._makeContextArgs(targetObject, '_init', initResolve, reject);
            targetObject['_init'](initContextArgs);
            return;
          }
        }

        that._callAction(targetObject, actionName, actionContextArgs, actionArgs);
      });
    }
  }, {
    key: '_makeContextArgs',
    value: function _makeContextArgs(targetObject, actionName, resolve, reject) {
      var contextArgs = { target: targetObject, action: actionName, actions: this, resolve: resolve, reject: reject };
      var _iteratorNormalCompletion = true;
      var _didIteratorError = false;
      var _iteratorError = undefined;

      try {
        for (var _iterator = this._contexts[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
          var context = _step.value;

          for (var key in context) {
            if (!contextArgs[key]) {
              contextArgs[key] = context[key];
              // console.log(key)
              // console.log(contextArgs)
            }
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

      return contextArgs;
    }
  }, {
    key: '_callAction',
    value: function _callAction(targetObject, actionName, actionContextArgs, actionArgs) {
      var _this = this;

      if (this._filters.length === 0) {
        targetObject[actionName](actionContextArgs, actionArgs);
        return;
      }

      var filterResolve = function filterResolve() {
        targetObject[actionName](actionContextArgs, actionArgs);
      };
      var filterContextArgs = this._makeContextArgs(targetObject, actionName, filterResolve, actionContextArgs.reject);

      var _loop = function _loop(i) {
        var filter = _this._filters[i];
        var nextFilterContextArgs = filterContextArgs;
        filterResolve = function filterResolve() {
          filter(nextFilterContextArgs, actionArgs);
        };
        filterContextArgs = _this._makeContextArgs(targetObject, actionName, filterResolve, actionContextArgs.reject);
      };

      for (var i = this._filters.length - 1; i >= 0; i--) {
        _loop(i);
      }
      filterResolve();
    }
  }]);
  return Action;
}();

function _cast(type, value) {
  var srcType = typeof value === 'undefined' ? 'undefined' : _typeof(value);
  switch (type) {
    case 'int':
      if (srcType === 'boolean') return value ? 1 : 0;
      if (!value) return 0;
      return parseInt(value);
    case 'float':
      if (srcType === 'boolean') return value ? 1.0 : 0.0;
      if (!value) return 0.0;
      return parseFloat(value);
    case 'boolean':
      if (srcType === 'string') return value === '0' || value === '' || value.toLowerCase() === 'false' ? false : true;
      return value ? true : false;
    case 'string':
      if (srcType === 'boolean') return value ? 'true' : 'false';
      if (srcType === 'string') return value;
      if (!value) return '';
      return value + '';
    case 'array':
      if (!(value instanceof Array)) {
        return [value];
      }
  }
  return value;
}
var _checkRegexCache = {};

var SimpleChecker = (function (_ref, args) {
  var target = _ref.target,
      action = _ref.action,
      resolve = _ref.resolve,
      reject = _ref.reject;


  var checkDefines = target['_' + action];
  if ((typeof checkDefines === 'undefined' ? 'undefined' : _typeof(checkDefines)) === 'object') {
    for (var field in checkDefines) {
      var define = checkDefines[field];
      var value = args[field];
      if (value === undefined) {
        if (define.require) {
          return reject(new Error(define.message || '[CHECK FAILED] ' + field + ' is requird'));
        }
      } else {
        if (define.type) {
          args[field] = _cast(define.type, value);
        }
        if (define.checker) {
          var checkOk = true;
          if (typeof define.checker === 'function') {
            if (!define.checker(args[field])) checkOk = false;
          } else {
            var strValue = _cast('string', value);
            if (!_checkRegexCache[define.checker]) _checkRegexCache[define.checker] = new RegExp(define.checker);
            checkOk = strValue.match(_checkRegexCache[define.checker]) !== null;
          }
          if (!checkOk) {
            return reject(new Error(define.message || '[CHECK FAILED] ' + field + ':[' + value + '] check failed with ' + define.checker));
          }
        }
      }
    }
  }
  resolve();
});

exports.Action = Action;
exports.SimpleChecker = SimpleChecker;

Object.defineProperty(exports, '__esModule', { value: true });

})));
