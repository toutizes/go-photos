/*global $, console, TT_Fetcher2, TT_Montage, TT_DB5, TT_Preloader, tt_Infinite, TT_Keywords*/
/*jslint browser:true, nomen: true, unparam: true*/
var TT_Image5 = (function () {
  "use strict";

  // Display elements.
  var images_ = null;		// Container for IMAGE_MODE.
  var mini_model_ = null;	// Model for minis.
  var mini_size_ = 0;		// Size of the mini model.
  var container_ = null;	// Container for minis.
  var prev_ = null; 		// Previous image button.
  var next_ = null; 		// Next image button.

  // Utility objects.
  var slider_ = null;		// TT_Slider for images.
  var preloader_ = null;	// TT_Preloader for images.
  var infinite_ = null;		// Infinite scroller object for minis.
  var montage_ = null;		// TT_Montage for mini images.

  // Data.
  var h_ = null;		// Hash currently displayed.
  var h_images_ = null;		// Images queried from the hash.
  var cur_mini_ = null;		// Element for current mini.

  function image_album_dir(image) {
    var album_dir = image.albumId;
    if (album_dir.slice(0, 10) === "originals/") {
      album_dir = album_dir.slice(10);
    }
    return album_dir;
  }

  function image_midi_url(image) {
    return "midi/" + image_album_dir(image) + "/" + image.filename;
  }

  function image_maxi_url(image) {
    return "maxi/" + image_album_dir(image) + "/" + image.filename;
  }

  function req_mini(e) {
    var data = $(this).data();
    TT_DB5.req_image_index(data.index);
  }

  function preload_next_images(images, c) {
    if (images === h_images_) {
      if (c + 1 < images.length) {
        preloader_.preload(image_midi_url(images[c + 1]));
      }
    }
  }

  function make_mini_div(idx) {
    var index = 3 * idx;
    if (index < 0 || index >= h_images_.length) {
      return null;
    }
    var row = $("<div/>", {"class": "images-mini-row"});
    var i, mini;
    for (i = index; i < index + 3 && i < h_images_.length; i++) {
      mini = mini_model_.clone();
      mini.attr("id", "m-" + i);
      mini.data({index: i});
      mini.css(montage_.bg_style(i));
      mini.click(req_mini);
      row.append(mini);
    }
    return row;
  }

  function minis_per_page() {
    var n = Math.ceil((1 + (container_.width() / (3 * mini_size_))) *
                      (container_.height() / mini_size_));
    return n;
  }

  function stereo_if_needed(h, photo) {
    if (h.stereo && !photo.hasOwnProperty("stereo_dx")) {
      photo.stereo_dx = 0.0;
      photo.stereo_dy = 0.0;
    }
    return photo;
  }

  function show_midi(h, images) {
    var photo, img_url, cb;
    if (images.length > 0 && h.c < images.length) {
      cb = function () { preload_next_images(images, h.c); };
      photo = stereo_if_needed(h, images[h.c]);
      img_url = image_midi_url(photo);
      if (h_ === null || h_.q !== h.q || h.c === h_.c) {
        slider_.center(h, photo, img_url, cb);
      } else if (h.c > h_.c) {
        slider_.slide_left(h, photo, img_url, cb);
      } else {
        slider_.slide_right(h, photo, img_url, cb);
      }
    } else {
      slider_.clear();
    }
  }

  function show_maxi(h, images) {
    var photo, img_url, cb;
    if (images.length > 0 && h.c < images.length) {
      cb = function () { preload_next_images(images, h.c); };
      photo = stereo_if_needed(h, images[h.c]);
      img_url = image_midi_url(photo);
      slider_.show_full(h, photo, img_url, cb);
    } else {
      slider_.clear();
    }
  }

  function show_download_links(h, images) {
    var image, url;
    if (images.length > 0 && h.c < images.length) {
      image = images[h.c];
      url = "viewer?command=download&q=" + encodeURIComponent($("#q").val());
      $("#nav-download").attr("href", image_maxi_url(image) + "?dl=true");
      $("#nav-download").attr("download", image.filename);
      $("#nav-download-all").attr("href", url + "&s=L");
      $("#nav-download-all-small").attr("href", url + "&s=M");
      $("#download-area").css("visibility", "visible");
    } else {
      $("#download-area").css("visibility", "hidden");
    }
  }

  function fix_h(h, images) {
    var i;
    if (h.k !== null) {
      // Set h.c to the first images with id h.k.  If no image is
      // found h.c is set to 0.
      h.c = 0;
      for (i = 0; i < images.length; i++) {
	// NOTE: use '==' and not '===' as h.k and image.id can be of different types.
	if (images[i].id == h.k) {
	  h.k = null;
	  h.c = i;
	  break;
	}
      }
    }
    if (h.c >= images.length) {
      h.c = images.length - 1;
    }
    if (h.c < 0) {
      h.c = 0;
    }
  }

  function images_ready(h, images) {
    var ids = $.map(images, function (i) { return i.id; });
    montage_ = TT_Montage.create(8, Math.floor(mini_size_), ids);
    fix_h(h, images);
    if (h.c === 0) {
      prev_.addClass("hidden");
    } else {
      prev_.removeClass("hidden");
    }
    if (h.c === images.length - 1) {
      next_.addClass("hidden");
    } else {
      next_.removeClass("hidden");
    }
    if (h.full) {
      images_.addClass("muted");
    } else {
      images_.removeClass("muted");
    }
    if (h.full) {
      show_maxi(h, images);
    } else {
      show_midi(h, images);
    }
    var rebuild_minis = (h_images_ === null || images !== h_images_);
    h_ = h;
    h_images_ = images;
    if (infinite_ === null) {
      infinite_ = make_infinite();
    }
    if (rebuild_minis) {
      infinite_.display(Math.floor(h_.c / 3));
    }
    // highlight current mini, and make it visible
    if (cur_mini_) {
      cur_mini_.removeClass("mini-current");
    }
    cur_mini_ = $("#m-" + h_.c).eq(0);
    if (cur_mini_) {
      cur_mini_.addClass("mini-current");
      infinite_.scroll_into_view(Math.floor(h_.c / 3));
    }
    TT_Keywords.display(h_images_[h_.c]);
    show_download_links(h_, h_images_);
  }

  function display(h) {
    if (h_ === null || h_.q !== h.q) {
      TT_Fetcher2.queryPhotos(h.q, function (images) { images_ready(h, images); });
    } else {
      images_ready(h, h_images_);
    }
  }

  function hide() {
    images_.addClass("hidden");
    $("#download-area").css("visibility", "hidden");
  }

  function show() {
    images_.removeClass("hidden");
  }

  function horizontal() {
    return container_.width() > container_.height();
  }

  function make_infinite() {
    return tt_Infinite(container_, $("#mini-contents"), horizontal(),
    		       { items_per_page: minis_per_page,
    			 make_item_div: make_mini_div
    		       });
  }

  function resized() {
    if (slider_ !== null) {
      slider_.resized();
    }
    if (infinite_ !== null) {
      if (infinite_.horizontal() !== horizontal()) {
	infinite_.destroy();
	infinite_ = make_infinite();
	if (h_ !== null) {
	  infinite_.display(Math.floor(h_.c / 3));
	  cur_mini_ = $("#m-" + h_.c).eq(0);
	  if (cur_mini_ !== null) {
	    cur_mini_.addClass("mini-current");
	  }
	}
      } else if (h_ !== null) {
	infinite_.scroll_into_view(Math.floor(h_.c / 3));
      }
    }
  }

  function initialize(slider) {
    images_ = $("#images");
    // keyword_model_ = $("#keyword-model");
    // keywords_ = $("#keywords");
    mini_model_ = $("#mini-model");
    mini_size_ = mini_model_.width();
    container_ = $("#mini-container");
    prev_ = $(".prev");
    next_ = $(".next");

    slider_ = slider;
    preloader_ = TT_Preloader.create(10);
  }

  return {
    resized: resized,
    display: display,
    hide: hide,
    show: show,
    initialize: initialize,
    // For iflow
    fix_h: fix_h
  };
}());
