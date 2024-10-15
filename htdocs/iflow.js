/*global $, console, TT_Fetcher2, TT_Image5, TT_DB5, tt_Infinite, TT_Keywords, TT_Preloader*/
/*jslint browser:true, nomen: true, unparam: true*/
var TT_IFlow = (function () {
  "use strict";

  var target_h_ = 300;		// Target height of each row.
  var container_ = null;	// Iflow div.
  var prev_ = null; 		// Previous image button.
  var next_ = null; 		// Next image button.
  var img_model_ = null;	// Model for images.
  var image_count_= null;       // Element that displays the count of images.

  var h_ = null;		// last hash used for display.
  var h_images_ = null;		// last images fetched from h_.
  var pad_ = 0;			// Pad around row images.

  // Precomputed indices of row starts.  Maps a display row index to
  // an index in h_images_.  Built by recompute_row_indices().
  var row_indices_ = null;
  var infinite_ = null;		// The infinite scroller.

  var slider_ = null;		// Used to display full size images.
  var preloader_ = null;	// TT_Preloader for images.

  function req_full(e) {
    var data = $(this).data();
    TT_DB5.req_full(data.index);
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
      if (img.w === 0 || img.h === 0) {
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
      img_elt.click(req_full);
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

  function clear_infinite() {
    if (infinite_ !== null) {
      infinite_.destroy();
      infinite_ = null;
    }
  }

  function preload_next_images(images, c) {
    return function() {
      if (images === h_images_) {
	if (c + 1 < images.length) {
          preloader_.preload(image_img_url(images[c + 1], "midi"));
	}
      }
    };
  }

  function show_download_links(h, images) {
    var image, url;
    if (images.length > 0 && h.c < images.length) {
      image = images[h.c];
      url = "viewer?command=download&q=" + encodeURIComponent($("#q").val());
      $("#nav-download").attr("href", image_img_url(image, "maxi") + "?dl=true");
      $("#nav-download").attr("download", image.filename);
      $("#nav-download-all").attr("href", url + "&s=L");
      $("#nav-download-all-small").attr("href", url + "&s=M");
      $("#download-area").css("visibility", "visible");
    } else {
      $("#download-area").css("visibility", "hidden");
    }
  }

  function images_ready(h, images) {
    image_count_.text("Images: " + images.length);
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
      console.log("full");
      console.log(h.c);
      console.log(images.length);
      if (h.c === 0) {
        console.log("hiding prev");
        prev_.addClass("hidden");
      } else {
        prev_.removeClass("hidden");
      }
      if (h.c === images.length - 1) {
        next_.addClass("hidden");
      } else {
        next_.removeClass("hidden");
      }
      container_.addClass("muted");
    } else {
      prev_.addClass("hidden");
      next_.addClass("hidden");
      container_.removeClass("muted");
    }
    $("#" + image_eltid(h_.c)).addClass("iflow-img-current");
    var img = h_images_[h_.c];
    if (h.full) {
      slider_.show_full(h_, img, image_img_url(img, "midi"), 
			preload_next_images(images, h.c));
      slider_.show();
    } else {
      slider_.hide();
    }
    TT_Keywords.display(h_images_[h_.c]);
    show_download_links(h_, h_images_);
  }

  function display(h) {
    if (h_ === null || h_.q !== h.q) {
      clear_infinite();
      TT_Fetcher2.queryPhotos(h.q, function (images) { images_ready(h, images); });
    } else {
      images_ready(h, h_images_);
    }
  }

  function resized() {
    if (slider_ !== null) {
      slider_.resized();
    }
    clear_infinite();
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
    prev_ = $("#prev-full");
    next_ = $("#next-full");
    img_model_ = $("#iflow-img-model");
    image_count_ = $("#image-count");
    slider_ = slider;
    preloader_ = TT_Preloader.create(10);
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
