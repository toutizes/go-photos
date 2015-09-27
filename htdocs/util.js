/*global $*/
/*jslint browser: true*/
var TT_Util = (function () {
  "use strict";
  function elementHeight(html) {
    var elt, result;
    elt = $(html);
    $("body").append(elt);
    result = elt.outerHeight();
    elt.remove();
    return result;
  }

  function elementBorder(e) {
    return e.outerWidth() - e.width();
  }

  return {
    elementHeight: elementHeight,
    elementBorder: elementBorder
  };
}());
