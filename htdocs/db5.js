/*global $, console, TT_Album5, TT_Image5, TT_Hash, TT_DisplayMode2, TT_ThreeD2, TT_Slider5, TT_IFlow, TT_Keywords*/
/*jslint browser:true, nomen: true*/
var TT_DB5 = (function () {
  "use strict";

  var h_ = null,
    content_ = null,
    portrait_ = null,
    header_,
    q_,
    last_image_mode_ = "F",
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

  function set_hash(h, push_history) {
    h = TT_Hash.fill_defaults(h, TT_Hash.get());
    TT_Hash.set(h, push_history);
  }

  function resized() {
    var portrait = $(window).height() > $(window).width(),
      m = $("#mini-container"),
      sb;
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
      TT_Keywords.show();
      TT_Album5.hide();
      TT_IFlow.hide();
      content_ = TT_Image5;
    }
  }

  function display_iflow(h) {
    TT_IFlow.display(h);
    if (content_ !== TT_IFlow) {
      TT_IFlow.show();
      TT_Keywords.show();
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
      TT_Keywords.hide();
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
    h_ = get_hash();
    display_form(h_);
    // p2 is backwards compatibility for redirects from feed that goes through
    // db/s.  Change db/s to use mode "I"!
    if (h_.mode === IMAGE_MODE || h_.mode === "p2") {
      display_images(h_);
    } else if (h_.mode === IFLOW_MODE) {
      display_iflow(h_);
    } else if (h_.mode === ALBUM_MODE || h_.mode === NEWS_MODE) {
      display_albums(h_);
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
    set_hash({mode: last_image_mode_}, true);
  }

  function req_iflow_mode() {
    set_hash({mode: IFLOW_MODE}, true);
  }

  function req_image_index(i) {
    set_hash({c: i, k: null}, true);
  }

  function req_midi_images() {
    set_hash({full: false}, false);
  }

  function req_full(i) {
    set_hash({c: i, k: null, full: true}, false);
  }

  function toggle_midi_maxi_images() {
    var h = get_hash();
    h.full = !h.full;
    set_hash(h, false);
  }

  function toggle_iflow() {
    var h = get_hash();
    if (h_.mode === IMAGE_MODE) {
      h.mode = IFLOW_MODE;
      last_image_mode_ = h.mode;
      set_hash(h, false);
    } else if (h_.mode === IFLOW_MODE) {
      h.mode = IMAGE_MODE;
      last_image_mode_ = h.mode;
      set_hash(h, false);
    }
  }

  function toggle_stereo() {
    var h = get_hash();
    h.stereo = !h.stereo;
    set_hash(h, false);
  }

  function stereo_mode(st_mode) {
    var h = get_hash();
    if (st_mode) {
      h.stereo = true;
      h.st_mode = st_mode.code;
    } else {
      h.stereo = false;
    }
    set_hash(h, false);
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
    set_hash({mode: ALBUM_MODE, stereo: false}, true);
  }

  function req_string(q, k) {
    console.log('req_string ' + q + ', ' + k);
    var mode = (h_.mode == ALBUM_MODE || h_.mode == NEWS_MODE) ? last_image_mode_ : h_.mode;
    var h = {mode: mode, q: q, stereo: false, full: false};
    if (k === null) {
      h.c = 0;
      h.k = null;
    } else {
      h.c = null;
      h.k = k;
    }
    set_hash(h, true);
  }

  function search() {
    console.log('search ' + q_.val());
    set_hash({mode: last_image_mode_, q: q_.val(), c: 0, k: null, stereo: false}, true);
  }

  function rel_next(n) {
    if (h_ !== null) {
      set_hash({c: Math.max(0, h_.c + n), k: null}, false);
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
    case 32:                    // space
      event.preventDefault();
      toggle_midi_maxi_images();
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
    $("#midi-model").click(toggle_midi_maxi_images);
    $("#iflow-img-model").dblclick(toggle_midi_maxi_images);
    $(document.body).keydown(key_down);
    $("#q").keydown(function (e) {
      e.stopPropagation();
      if (e.keyCode === 13) {
        search();
      }
    });
  }

  function onSignIn(googleUser) {
    console.log("onSignin");
    console.log(googleUser);
    var id_token = googleUser.getAuthResponse().id_token;
    // Send the ID token to your backend for verification
    $.ajax({
      type: 'POST',
      url: '/auth-google', // Your backend endpoint
      data: { id_token: id_token },
      success: function(response) {
        // TODO: redirect to authpic.html.
        // Handle successful authentication (e.g., redirect, update UI)
        console.log("Success");
        console.log(response);
      },
      error: function(error) {
        // Handle authentication error
        console.log("Failure");
        console.log(error);
      }
    });
  }

  function initialize() {
    resized();
    bind_events();
    TT_ThreeD2.initialize();
    TT_Slider5.initialize(TT_ThreeD2);
    TT_Image5.initialize(TT_Slider5);
    TT_IFlow.initialize(TT_Slider5);
    TT_Album5.initialize();
    TT_Keywords.initialize();
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
    req_album_mode: req_album_mode,
    req_image_index: req_image_index,
    req_string: req_string,
    req_full: req_full,
    st_mode: st_mode,
    search: search,
    iflow: toggle_iflow,
    onSignIn: onSignIn
  };
}());
