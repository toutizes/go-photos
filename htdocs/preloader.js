/*jslint browser: true*/
var TT_Preloader = {
  next: 0,
  images: [],

  preload: function (url) {
    'use strict';
    console.log('Preload: ' + url);
    this.images[this.next].src = url;
    this.next = (this.next + 1) % this.images.length;
  },

  create: function (max_preloads) {
    'use strict';
    var o, i;
    o = Object.create(this);
    for (i = 0; i < max_preloads; i++) {
      o.images.push(new Image());
    }
    return o;
  }
};
