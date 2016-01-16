/*global $, console, TT_Fetcher2, TT_Montage, TT_DB5, TT_Preloader*/
/*jslint browser:true, nomen: true, unparam: true*/
var TT_IFlow = (function () {
  "use strict";

  var h_ = null;  // last hash used for display.
  var h_images_ = null;  // last images fetched from h_.
  var iflow_ = null;  // iflow div.
  var preloader_ = null;  // TT_Preloader object.
  var pad_ = 0;  // Pad around row images.

  function image_img_url(image, kind) {
    return kind + "/" + image.albumId + "/" + image.filename;
  }

  function padded(x) {
    return x + pad_;
  }
  function unpadded(x) {
    return x - pad_;
  }

  function images_ready(h, images) {
    var i = 0, row = [], target_h = 200, target_w = iflow_.innerWidth(), w = 0;
    var i_w = 0, row_h = 0, j = 0, row_div = null;
    iflow_.empty();
    i = 0;
    while (i < images.length) {
      w = 0;
      row = [];
      while (w < target_w && i < images.length) {
	i_w = padded((images[i].w * target_h) / images[i].h);
	if (row.length > 0 && w + (i_w / 2) > target_w) {
	  break;
	}
	row.push(images[i]);
	i = i + 1;
	w = w + i_w;
      }
      if (row.length > 0) {
	if (i < images.length) {
	  row_h = unpadded((target_h * target_w) / w);
	} else {
	  row_h = unpadded(target_h);
	}
	row_div = $("<div/>", {class: "flow_row" });
	for (j = 0; j < row.length; j++) {
	  row_div.append($("<img/>", {
		class: "flow_img",
		src: image_img_url(row[j], "mini"),
		  height: row_h,
		  width: (row[j].w * row_h) / row[j].h
		  }));
	}
	iflow_.append(row_div);
      }
    }
  }

  function display(h) {
    TT_Fetcher2.queryPhotos(h.q, function (images) { images_ready(h, images); });
  }

  function resized() {
  }

  function hide() {
    iflow_.addClass("hidden");
  }

  function show() {
    iflow_.removeClass("hidden");
  }

  function initialize() {
    iflow_ = $("#iflow");
    preloader_ = TT_Preloader.create(10);
    pad_ = 10;
  }

  return {
    resized: resized,
    display: display,
    hide: hide,
    show: show,
    initialize: initialize
  };
}());
