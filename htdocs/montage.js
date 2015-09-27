/*global $, console*/
/*jslint browser: true, unparam: true, plusplus: true, regexp: true*/
var TT_Montage = (function () {
  "use strict";

  // Class attributes and methods.
  var MONTAGE_URL;
  MONTAGE_URL = "montage/";

  function build_montages(o) {
    var i, m, ms;
    m = null;
    ms = [];
    for (i = 0; i < o.names.length; ++i) {
      if (i % o.batch === 0) {
        if (m !== null) {
          ms.push(m);
        }
        m = MONTAGE_URL + o.size + "x" + o.size;
      }
      m += "-" + o.names[i];
    }
    ms.push(m);
    o.montages = ms;
  }

  function bg_style(o, index) {
    return { "background-image": "url('" + o.montages[Math.floor(index / o.batch)]  + "')",
             "background-position": "-" + (o.size * (index % o.batch)) + "px 0px" };
  }

  function img(o, index) {
    var image = $(new Image());
    image.attr("src", o.montages[Math.floor(index / o.batch)]);
    image.css({margin: "0px 0px 0px -" + (o.size * (index % o.batch)) + "px"});
    return image;
  }

  function create(o, batch, size, names) {
    o.batch = batch;
    o.size = size;
    o.names = names;
    build_montages(o);
    return o;
  }

  return {
    // Attributes
    size: 0,
    names: [],
    montages: [],

    // Methods
    bg_style: function (index) { return bg_style(this, index); },
    img: function (index) { return img(this, index); },
    create: function (batch, size, names) {
      return create(Object.create(this), batch, size, names);
    }
  };
}());
