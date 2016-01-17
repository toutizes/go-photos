/*global $, console, TT_Album5, TT_Image5, TT_Hash, TT_DisplayMode2, TT_ThreeD2, TT_Slider5, TT_IFlow*/
/*jslint browser:true, nomen: true*/
var TT_DB5 = (function () {
  "use strict";

  var content_ = null,
    portrait_ = null,
    header_,
    q_,
    IMAGE_MODE,
    IFLOW_MODE,
    ALBUM_MODE,
    NEWS_MODE;

  IMAGE_MODE = "I";
  IFLOW_MODE = "F";
  ALBUM_MODE = "A";
  NEWS_MODE = "N";

  // Hash management.
  function get_hash() {
    // var h = TT_Hash.get_with_defaults({mode: IMAGE_MODE,
    //                                    q: "catherine julien coline matthieu",
    //                                    c: 0,
    //                                    k: null,
    //                                    full: false});
    var h = TT_Hash.get_with_defaults({mode: ALBUM_MODE,
                                       q: "",
                                       c: 0,
                                       k: null,
                                       full: false,
                                       stereo: false,
                                       st_mode: TT_DisplayMode2.ACTIVE.code});
    h.c = Math.max(0, parseInt(h.c, 10));
    h.st_mode = parseInt(h.st_mode, 10);
    return h;
  }

  function set_hash(h) {
    var push_history = false;
    h = TT_Hash.fill_defaults(h, TT_Hash.get());
    TT_Hash.set(h, push_history);
  }

  function resized() {
    var portrait = $(window).height() > $(window).width(),
      m = $("#mini-container"),
      sb;
    // Update scrollbars on the mini container.
    if (portrait_ !== portrait) {
      if (portrait_ !== null) {
        // Remove previous scrollbars on minis.
        m.mCustomScrollbar("destroy");
      }
      if (portrait) {
        sb = {scrollInertia: 0,
              horizontalScroll: true,
              scrollButtons: {enable: true},
              advanced: {autoExpandHorizontalScroll: true},
              callbacks: {onTotalScroll: TT_Image5.append_more_minis,
                          onTotalScrollOffset: 3 * 64,
                          onTotalScrollBack: TT_Image5.prepend_more_minis,
                          onTotalScrollBackOffset: 3 * 64}};
      } else {
        sb = {scrollInertia: 0,
              scrollButtons: {enable: true},
              callbacks: {onTotalScroll: TT_Image5.append_more_minis,
                          onTotalScrollOffset: 3 * 64,
                          onTotalScrollBack: TT_Image5.prepend_more_minis,
                          onTotalScrollBackOffset: 3 * 64 }};
      }
      m.mCustomScrollbar(sb);
    }
    // Tell contents about the size change.
    if (content_ !== null) {
      content_.resized();
    }
    // Remember the new format.
    portrait_ = portrait;
  }

  // Responses to hash changes.
  function display_images(h) {
    TT_Image5.display(h);
    if (content_ !== TT_Image5) {
      TT_Image5.show();
      TT_Album5.hide();
      TT_IFlow.hide();
      content_ = TT_Image5;
    }
  }

  function display_iflow(h) {
    TT_IFlow.display(h);
    if (content_ !== TT_IFlow) {
      TT_IFlow.show();
      TT_Album5.hide();
      TT_Image5.hide();
      content_ = TT_IFlow;
    }
  }

  function display_albums(h) {
    $("#controls3d").addClass("hidden");
    TT_Album5.display(h);
    if (content_ !== TT_Album5) {
      TT_Album5.show();
      TT_Image5.hide();
      TT_IFlow.hide();
      content_ = TT_Album5;
    }
  }

  function display_form(h) {
    if (h.full) {
      header_.addClass("muted");
    } else {
      header_.removeClass("muted");
    }
    q_.val(h.q);
  }

  function hash_change() {
    var h = get_hash();
    display_form(h);
    // p2 is backwards compatibility for redirects from feed that goes through
    // db/s.  Change db/s to use mode "I"!
    if (h.mode === IMAGE_MODE || h.mode === "p2") {
      display_images(h);
    } else if (h.mode === IFLOW_MODE) {
      display_iflow(h);
    } else if (h.mode === ALBUM_MODE || h.mode === NEWS_MODE) {
      display_albums(h);
    }
  }

  function keywords_show_hide() {
    $(this).toggleClass("keywords-hide");
    $(this).toggleClass("keywords-show");
    if ($(this).hasClass("keywords-show")) {
      $("#keywords-container").css({ "max-width": "8em", "max-height": "23px"});
    } else {
      $("#keywords-container").css({"max-width": "20em", "max-height": "20ex"});
    }
  }

  // Requests for hash change.
  function req_image_mode() {
    set_hash({mode: IMAGE_MODE});
  }

  function req_iflow_mode() {
    set_hash({mode: IFLOW_MODE});
  }

  function req_image_index(i) {
    set_hash({c: i, k: null});
  }

  function req_midi_images() {
    set_hash({full: false});
  }

  function toggle_midi_maxi_images() {
    var h = get_hash();
    h.full = !h.full;
    set_hash(h);
  }

  function toggle_stereo() {
    var h = get_hash();
    h.stereo = !h.stereo;
    set_hash(h);
  }

  function stereo_mode(st_mode) {
    var h = get_hash();
    if (st_mode) {
      h.stereo = true;
      h.st_mode = st_mode.code;
    } else {
      h.stereo = false;
    }
    set_hash(h);
  }

  function st_mode(name) {
    if (name === 'active') {
      stereo_mode(TT_DisplayMode2.ACTIVE);
    } else if (name === 'parallel') {
      stereo_mode(TT_DisplayMode2.PARALLEL);
    } else if (name === 'crosseye') {
      stereo_mode(TT_DisplayMode2.CROSSEYE);
    }
  }

  function req_album_mode() {
    set_hash({mode: ALBUM_MODE, stereo: false});
  }

  function req_string(q, k) {
    set_hash({mode: IMAGE_MODE, q: q, c: 0, k: k, stereo: false});
  }

  function search() {
    set_hash({mode: IMAGE_MODE, q: q_.val(), c: 0, k: null, stereo: false});
  }

  function iflow() {
    set_hash({mode: IFLOW_MODE, q: q_.val(), c: 0, k: null, stereo: false});
  }

  function rel_next(n) {
    var h = TT_Image5.h();
    if (h !== null) {
      set_hash({c: Math.max(0, h.c + n), k: null});
    }
  }

  function next() {
    rel_next(1);
  }

  function prev() {
    rel_next(-1);
  }

  function key_down(event) {
    TT_ThreeD2.key_down(event);
    switch (event.which) {
    case 37:                        // left arrow
      event.preventDefault();
      rel_next(-1);
      break;
    case 39:                        // right arrow
      event.preventDefault();
      rel_next(1);
      break;
    case 38:                        // up arrow
      event.preventDefault();
      rel_next(-3);
      break;
    case 40:                        // down arrow
      event.preventDefault();
      rel_next(3);
      break;
    case 27:                        // escape
      event.preventDefault();
      req_midi_images();
      break;
    case 83:                    // S
      event.preventDefault();
      toggle_stereo();
      break;
    case 80:                    // P
      event.preventDefault();
      stereo_mode(TT_DisplayMode2.PARALLEL);
      break;
    case 88:                    // X
      event.preventDefault();
      stereo_mode(TT_DisplayMode2.CROSSEYE);
      break;
    case 65:                    // A
      event.preventDefault();
      stereo_mode(TT_DisplayMode2.ACTIVE);
      break;
    case 79:                    // O
      event.preventDefault();
      stereo_mode(TT_DisplayMode2.OVERLAY);
      break;
    case 68: // D
    case 87: // W
      event.preventDefault();
      // Follows the 'D' and 'W' from 3d2.js.
      stereo_mode(null);
      break;
    }
  }

  function bind_events() {
    $(window).hashchange(hash_change);
    if ($(window).orientation !== undefined) {
      $(window).bind("orientationchange", resized);
    } else {
      $(window).resize(resized);
    }
    $("#keywords-show-hide").click(keywords_show_hide);
    $("#prev").click(prev);
    $("#next").click(next);
    $("#prev-full").click(prev);
    $("#next-full").click(next);
    $("#midi-model").dblclick(toggle_midi_maxi_images);
    $(document.body).keydown(key_down);
    $("#q").keydown(function (e) {
      e.stopPropagation();
      if (e.keyCode === 13) {
        search();
      }
    });
  }

  function initialize() {
    resized();
    bind_events();
//     $.loading({onAjax: true, delay: 300, text: "Chargement..."});
    TT_ThreeD2.initialize();
    TT_Slider5.initialize(TT_ThreeD2);
    TT_Image5.initialize(TT_Slider5);
    TT_IFlow.initialize();
    TT_Album5.initialize();
    header_ = $("#header");
    q_ = $("#q");
    // Do it.
    hash_change();
  }

  function init() {
    $(document).ready(initialize);
  }

  return {
    init: init,
    req_image_mode: req_image_mode,
    req_iflow_mode: req_iflow_mode,
    req_album_mode: req_album_mode,
    req_image_index: req_image_index,
    req_string: req_string,
    st_mode: st_mode,
    search: search,
    iflow: iflow
  };
}());
