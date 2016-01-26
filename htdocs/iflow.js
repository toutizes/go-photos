/*global $, console, TT_Fetcher2, TT_Image5, TT_DB5, tt_Infinite, TT_Keywords*/
/*jslint browser:true, nomen: true, unparam: true*/
var TT_IFlow = (function () {
  "use strict";

  var target_h_ = 200;		// Target height of each row.
  var container_ = null;	// Iflow div.
  var img_model_ = null;	// Model for images.

  var h_ = null;		// last hash used for display.
  var h_images_ = null;		// last images fetched from h_.
  var pad_ = 0;			// Pad around row images.

  // Precomputed indices of row starts.  Maps a display row index to
  // an index in h_images_.  Built by recompute_row_indices().
  var row_indices_ = null;
  var infinite_ = null;		// The infinite scroller.

  var slider_ = null;		// Used to display full size images.

  function req_full(e) {
    var data = $(this).data();
    TT_DB5.req_full(data.index);
  }

  function req_image(e) {
    var data = $(this).data();
    TT_DB5.req_image_index(data.index);
  }

  function image_img_url(image, kind) {
    return kind + "/" + image.albumId + "/" + image.filename;
  }

  function image_eltid(image_idx) {
    return "i-" + image_idx;
  }

  function padded(x) {
    return x + pad_;
  }
  function unpadded(x) {
    return x - pad_;
  }

  function recompute_row_indices() {
    var target_w = container_.innerWidth();
    var w_remaining = target_w;
    var i;
    row_indices_ = [0];
    var img, w_i;
    for (i = 0; i < h_images_.length; i++) {
      img = h_images_[i];
      if (img.w == 0 || img.h == 0) {
	console.log("No size: " + img.albumId + "/" + img.filename);
	img.w = 1200;
	img.h = 900;
      }
      w_i = padded((img.w * target_h_) / img.h);
      if (w_i / 2 > w_remaining) {
	row_indices_.push(i);
	w_remaining = target_w;
      } else {
	w_remaining -= w_i;
      }
    }
  }

  function image_index_to_row_index(image_index) {
    if (row_indices_ === null || image_index < 0) {
      return 0;
    }
    // Could do binary search.
    var i;
    for (i = row_indices_.length - 1; i >= 0; i--) {
      if (row_indices_[i] <= image_index) {
	return i;
      }
    }
    return 0;
  }

  function rows_per_page() {
    var n = 1 + Math.ceil(container_.height() / target_h_);
    return n;
  }

  function make_row_div(idx) {
    if (row_indices_ === null) {
      return null;
    }
    var from_i = row_indices_[idx];
    if (from_i === undefined) {
      return null;
    }
    var to_i = row_indices_[idx + 1];
    if (to_i === undefined) {
      to_i = h_images_.length;
    }
    var original_w = 0;
    var i, img;
    for (i = from_i; i < to_i; i++) {
      img = h_images_[i];
      original_w += padded((img.w * target_h_) / img.h);
    }
    var row_div = $("<div/>", {class: "flow_row" });
    var target_w = container_.innerWidth();
    var row_h = Math.floor(unpadded(target_h_ * (target_w / original_w)));
    var img_elt = null;
    if (idx === row_indices_.length - 1) {
      row_h = Math.min(row_h, target_h_);
    }
    for (i = from_i; i < to_i; i++) {
      img = h_images_[i];
      img_elt = img_model_.clone();
      img_elt.attr({
	id: image_eltid(i),
	src: image_img_url(img, "mini"),
	height: row_h
      });
      img_elt.data({index: i});
      img_elt.dblclick(req_full);
      img_elt.click(req_image);
      row_div.append(img_elt);
    }
    return row_div;
  }

  function make_infinite() {
    recompute_row_indices();
    infinite_ = tt_Infinite(container_, $("#iflow-contents"), 
			    false /* horizontal */,
    			    { items_per_page: rows_per_page,
    			      make_item_div: make_row_div
    			    });
  }

  function images_ready(h, images) {
    TT_Image5.fix_h(h, images);
    if (h_ !== null) {
      $("#" + image_eltid(h_.c)).removeClass("iflow-img-current");
    }
    var rebuild = (images !== h_images_);
    h_images_ = images;
    h_ = h;
    if (rebuild) {
      make_infinite();
    }
    var row_index = image_index_to_row_index(h.c);
    if (rebuild) {
      infinite_.display(row_index);
    } else {
      infinite_.scroll_into_view(row_index);
    }
    if (h.full) {
      container_.addClass("muted");
    } else {
      container_.removeClass("muted");
    }
    $("#" + image_eltid(h_.c)).addClass("iflow-img-current");
    var img = h_images_[h_.c];
    if (h.full) {
      slider_.show();
      slider_.show_full(h_, img, image_img_url(img, "midi"), null);
    } else {
      slider_.hide();
    }
    TT_Keywords.display(h_images_[h_.c]);
  }

  function display(h) {
    if (h_ === null || h_.q !== h.q) {
      TT_Fetcher2.queryPhotos(h.q, function (images) { images_ready(h, images); });
    } else {
      images_ready(h, h_images_);
    }
  }

  function resized() {
    if (slider_ !== null) {
      slider_.resized();
    }
    if (infinite_ !== null) {
      infinite_.destroy();
      infinite_ = null;
    }
    if (h_images_ !== null) {
      make_infinite();
      infinite_.display(0);
    }
  }

  function hide() {
    container_.addClass("hidden");
    slider_.hide();
  }

  function show() {
    container_.removeClass("hidden");
    if (h_ && h_.full) {
      slider_.show();
    }
  }

  function initialize(slider) {
    container_ = $("#iflow-container");
    img_model_ = $("#iflow-img-model");
    slider_ = slider;
    pad_ = 10;
    row_indices_ = [];
  }

  return {
    resized: resized,
    display: display,
    hide: hide,
    show: show,
    initialize: initialize
  };
}());
