/*global $, console, TT_Fetcher2, TT_Montage, TT_DB5, TT_Preloader*/
/*jslint browser:true, nomen: true, unparam: true*/
var TT_Image5 = (function () {
  "use strict";

  var images_, keywords_, mini_model_, mini_size_, keyword_model_, montage_,
    slider_ = null, prev_ = null, next_ = null, h_ = null, h_images_ = null,
    h_mini_from_ = null, h_mini_to_ = null, cur_mini_ = null,
    preloader_, MONTAGE = 8;

  function elt_visible(elt, cont) {
    var cont_top = cont.offset().top,
      cont_bot = cont_top + cont.height(),
      cont_left = cont.offset().left,
      cont_right = cont_left + cont.width(),
      elt_offest = elt.offset(),
      elt_top,
      elt_bot,
      elt_left,
      elt_right;
    if (elt_offest === null) {
      return false;
    }
    elt_top = elt.offset().top;
    elt_bot = elt_top + elt.outerHeight();
    elt_left = elt.offset().left;
    elt_right = elt_left + elt.outerWidth();
    return (elt_top >= cont_top && elt_bot <= cont_bot &&
            elt_left >= cont_left && elt_right <= cont_right);
  }

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

  function req_keyword(e) {
    var data = $(this).data();
    TT_DB5.req_string(data.kw, data.image);
  }

  function preload_next_images(images, c) {
    if (images === h_images_) {
      if (c + 1 < images.length) {
        preloader_.preload(image_midi_url(images[c + 1]));
      }
    }
  }

  function get_mini_container() {
    return $("#mini-container");
  }

  function get_mini_contents() {
    return $("#mini-container .mCSB_container");
  }

  function num_minis_in_viewport() {
    var container = get_mini_container();
    return Math.ceil((container.width() / mini_size_) *
                     (container.height() / mini_size_));
  }

  function make_mini(abs_index) {
    var mini = mini_model_.clone();
    mini.attr("id", "m-" + abs_index);
    mini.data({index: abs_index});
    mini.css(montage_.bg_style(abs_index));
    mini.click(req_mini);
    return mini;
  }

  function add_minis(images, mini_from, mini_to) {
    var mini_contents = get_mini_contents(),
      prepend = (h_mini_from_ === null || mini_from <= h_mini_from_),
      cur_mini_pos,
      delta,
      i;
    mini_from = Math.max(0, mini_from);
    mini_to = Math.min(images.length, mini_to);
    if (mini_to === mini_from) {
      return;
    }
//    console.log("add_minis ");
//    console.log(cur_mini_);
//    if (cur_mini_) {
//      console.log("add_minis_pos ");
//      console.log(cur_mini_.position());
//    }
    if (prepend) {
      if (cur_mini_) {
        cur_mini_pos = cur_mini_.position();
      }
      for (i = mini_to - 1; i >= mini_from; i--) {
        mini_contents.prepend(make_mini(i));
      }
    } else {
      for (i = mini_from; i < mini_to; i++) {
        mini_contents.append(make_mini(i));
      }
    }
    if (h_mini_from_ === null) {
      h_mini_from_ = mini_from;
      h_mini_to_ = mini_to;
    } else {
      h_mini_from_ = Math.min(h_mini_from_, mini_from);
      h_mini_to_ = Math.max(h_mini_to_, mini_to);
    }
    get_mini_container().mCustomScrollbar("update");
    if (prepend && cur_mini_ && cur_mini_.position()) {
      if (get_mini_container().height() > get_mini_container().width()) {
        delta = cur_mini_.position().top - cur_mini_pos.top;
        get_mini_container().mCustomScrollbar("scrollTo",
                                              mini_contents.position().top + delta,
                                              {scrollInertia: 0});
      } else {
        delta = cur_mini_.position().left - cur_mini_pos.left;
        get_mini_container().mCustomScrollbar("scrollTo",
                                              mini_contents.position().left + delta,
                                              {scrollInertia: 0});
      }
    }
  }

  function prepend_more_minis() {
    if (h_images_ !== null && h_mini_from_ !== null) {
      add_minis(h_images_, h_mini_from_ - num_minis_in_viewport(), h_mini_from_);
    }
  }

  function append_more_minis() {
    if (h_images_ !== null && h_mini_from_ !== null) {
      add_minis(h_images_, h_mini_to_, h_mini_to_ + num_minis_in_viewport());
    }
  }

  function show_minis(h, images) {
    var N = num_minis_in_viewport(), P = N * Math.floor(h.c / N);
    get_mini_contents().empty();
    h_mini_from_ = null;
    h_mini_to_ = null;
    add_minis(images, P - N, P + N);
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

  function add_one_keyword(index, keyword, keyword_q, image_id, container) {
    var kw = keyword_model_.clone().attr("id", null).text(keyword);
    if (keyword_q.indexOf(" ") !== -1) {
      keyword_q = "\"" + keyword_q + "\"";
    }
    kw.data({kw: keyword_q, image: image_id});
    kw.click(req_keyword);
    container.append(kw);
    container.append(" ");
  }

  function add_keyword(index, keyword, image_id, container) {
    add_one_keyword(index, keyword, keyword, image_id, container);
  }

  function show_keywords(h, images) {
    var image, fn;
    if (images.length > 0 && h.c < images.length) {
      image = images[h.c];
      keywords_.empty();
      $("#keywords-album").empty();
      add_one_keyword(0, image.albumId, "album:" + image.albumId, image.id,
                      $("#keywords-album"));
      fn = function (index, keyword) {
        add_keyword(index, keyword, image.id, keywords_);
      };
      $.each(image.keywords, fn);
    }
  }

  function set_hash_c(index, image, h) {
    if (image.id === h.k) {
      h.c = index;
      h.k = null;
      return false;
    }
  }

  function show_download_links(h, images) {
    var image, url;
    if (images.length > 0 && h.c < images.length) {
      image = images[h.c];
      url = "viewer?command=download&q=" + encodeURIComponent($("#q").val());
      $("#nav-download").attr("href", image_maxi_url(image) + "?dl=true");
      $("#nav-download-all").attr("href", url + "&s=L");
      $("#nav-download-all-small").attr("href", url + "&s=M");
      $("#download-area").css("visibility", "visible");
    } else {
      $("#download-area").css("visibility", "hidden");
    }
  }

  function by_timestamp_decreasing(a, b) {
    return b.timestamp - a.timestamp;
  }

  function by_timestamp_increasing(a, b) {
    return a.timestamp - b.timestamp;
  }

  function images_ready(h, images) {
    var names,
      rebuild_minis = false,
      mini_container = get_mini_container();
    if (h.q.match(/^album:/)) {
      images = images.sort(by_timestamp_increasing);
    } else {
      images = images.sort(by_timestamp_decreasing);
    }
    names = $.map(images, function (i) { return i.id; });
    montage_ = TT_Montage.create(MONTAGE, Math.floor(mini_size_), names);
    if (h.k !== null) {
      // Set h.c to the first images with id h.k.  If no image is
      // found h.c is set to 0.
      h.c = 0;
      $.each(images, function (index, image) { return set_hash_c(index, image, h); });
    }
    if (h.c >= images.length) {
      h.c = images.length - 1;
    }
    if (h.c < 0) {
      h.c = 0;
    }
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
    rebuild_minis = (h_images_ === null || images !== h_images_);
    if (rebuild_minis) {
      show_minis(h, images);
    }
    // highlight current mini, and make it visible
    if (cur_mini_) {
      // console.log("Old cur_mini_");
      // console.log(cur_mini_);
      cur_mini_.removeClass("mini-current");
    }
    cur_mini_ = $("#m-" + h.c).eq(0);
    // console.log("New cur_mini_");
    // console.log(cur_mini_);
    if (cur_mini_) {
      cur_mini_.addClass("mini-current");
      if (!elt_visible(cur_mini_, mini_container)) {
        mini_container.mCustomScrollbar("scrollTo", "#" + cur_mini_.attr("id"),
                                        {scrollInertia: rebuild_minis ? 0 : 250});
      }
    }
    show_keywords(h, images);
    show_download_links(h, images);
    h_ = h;
    h_images_ = images;
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

  function resized() {
    var mini_container = get_mini_container();
    if (slider_ !== null) {
      slider_.resized();
    }
    if (cur_mini_ && !elt_visible(cur_mini_, mini_container)) {
      mini_container.mCustomScrollbar("scrollTo", "#" + cur_mini_.attr("id"));
    }
  }

  function initialize(slider) {
    images_ = $("#images");
    keywords_ = $("#keywords");
    keyword_model_ = $("#keyword-model");
    mini_model_ = $("#mini-model");
    mini_size_ = mini_model_.width();
    slider_ = slider;
    prev_ = $(".prev");
    next_ = $(".next");
    preloader_ = TT_Preloader.create(10);
  }

  return {
    h: function () { return h_; },
    resized: resized,
    append_more_minis: append_more_minis,
    prepend_more_minis: prepend_more_minis,
    display: display,
    hide: hide,
    show: show,
    initialize: initialize
  };
}());
