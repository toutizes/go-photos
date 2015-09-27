/*global btoa, atob, console*/
/*jslint browser: true*/

// Hash management.
var TT_Hash = (function () {
  'use strict';
  var log = false;

  function hashEncode(o) {
    var e, k;
    if (log) {
      console.log('encode');
      console.log(o);
    }
    e = '';
    for (k in o) {
      if (o.hasOwnProperty(k)) {
        e += '\x00' + k + '\x00' + o[k];
      }
    }
    return btoa(e.substr(1));
  }

  function safe_atob(s) {
    try {
      return atob(s);
    } catch (e1) {
      try {
        console.log('+=');
        return atob(s + '=');
      } catch (e2) {
        console.log('+==');
        return atob(s + '==');
      }
    }
  }

  function hashDecode(s) {
    var o, a, i, v;
    o = {};
    a = safe_atob(s).split('\x00');
    for (i = 0; i < a.length - 1; i += 2) {
      v = a[i + 1];
      if (v === "null" || v === "") {
        v = null;
      } else if (v === "false") {
        v = false;
      } else if (v === "true") {
        v = true;
      } // else v is good as-is.
      o[a[i]] = v;
    }
    if (log) {
      console.log('decode');
      console.log(o);
    }
    return o;
  }

  function setHashWithHistory(h, push_history) {
    var hash, u, hash_index;
    hash = hashEncode(h);
    if (push_history) {
      window.location.hash = hash;
    } else {
      u = window.location.href;
      hash_index = u.indexOf('#');
      if (hash_index !== -1) {
        u = u.substring(0, hash_index);
      }
      window.location.replace(u + '#' + hash);
    }
  }

  function getHash() {
    var s;
    s = window.location.hash.substr(1);
    if (s.length === 0) {
      return {};
    }
    return hashDecode(s);
  }

  function fillDefaults(h, defaults) {
    var p;
    for (p in defaults) {
      if (defaults.hasOwnProperty(p)) {
        if (h[p] === undefined) {
          h[p] = defaults[p];
        }
      }
    }
    return h;
  }

  function getHashWithDefaults(defaults) {
    return fillDefaults(getHash(), defaults);
  }

  return {
    set: setHashWithHistory,
    get: getHash,
    fill_defaults: fillDefaults,
    get_with_defaults: getHashWithDefaults
  };
}());
