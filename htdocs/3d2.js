/*global $, console, TT_Fetcher2*/
/*jslint browser: true*/

var TT_DisplayMode2 = {
  PARALLEL: { code: 0, control: "p2-parallel" },
  CROSSEYE: { code: 1, control: "p2-crosseye" },
  ACTIVE: { code: 2, delay: 300, control: "p2-active" },
  // ANAGLYPH: { code: 3, control: "p2-anaglyph" },
  OVERLAY: { code: 4, control: "p2-overlay" },
  LEFT: { code: 5, control: "p2-left" }
};

var TT_ThreeD2 = (function () {
  'use strict';
  var max_parallel_dw_, mode_, div_, cur_img_, photo_, div_orig_width_,
    div_orig_height_, left_div_, right_div_, ticker_;

  function swap_sides() {
    if (right_div_ !== null && mode_ === TT_DisplayMode2.ACTIVE) {
      if (right_div_.css("display") === "none") {
        right_div_.show();
        left_div_.hide();
      } else {
        left_div_.show();
        right_div_.hide();
      }
      ticker_ = setTimeout(function () { swap_sides(); }, mode_.delay);
    }
  }

  function key_down(event) {
    var dw, dh, unit;
    if (!cur_img_) {
      return;
    }
    dw = div_orig_width_;
    dh = div_orig_height_;
    if (mode_ !== TT_DisplayMode2.ACTIVE && mode_ !== TT_DisplayMode2.OVERLAY) {
      unit = 0.0;
    } else if (event.shiftKey) {
      unit = 10.0;
    } else if (event.ctrlKey) {
      unit = 0.1;
    } else {
      unit = 1.0;
    }
    switch (event.which) {
    case 73: // I
      photo_.stereo_dy -= unit / (2 * dh);
      render();
      return;
    case 74: // J
      photo_.stereo_dx -= unit / (2 * dw);
      render();
      return;
    case 75: // K
      photo_.stereo_dy += unit / (2 * dh);
      render();
      return;
    case 76: // L
      photo_.stereo_dx += unit / (2 * dw);
      render();
      return;
    case 68: // D
      delete photo_.stereo_dx;
      delete photo_.stereo_dy;
      TT_Fetcher2.clearStereo(photo_.id);
      return;
    case 87: // W
      TT_Fetcher2.setStereo(photo_.id, photo_.stereo_dx, photo_.stereo_dy);
      return;
    default:
      return;
    }
  }

  function create_divs() {
    var dw, dh, iw, ih, s, url, h_port, vis_w, vis_h, x_off, y_off, style, contents;

    dw = div_orig_width_;
    dh = div_orig_height_;
    iw = cur_img_.width;
    ih = cur_img_.height;
    s = 1;


    url = cur_img_.src;
    h_port = 0.5;
    switch (mode_) {
    case TT_DisplayMode2.PARALLEL:
      s = Math.min(Math.min(dw, 2 * max_parallel_dw_) / iw, dh / ih);
      break;
    case TT_DisplayMode2.CROSSEYE:
      s = Math.min(dw / iw, dh / ih);
      break;
    case TT_DisplayMode2.ACTIVE:
    case TT_DisplayMode2.OVERLAY:
    case TT_DisplayMode2.LEFT:
      s = Math.min(2 * dw / iw, dh / ih);
      break;
    // case TT_DisplayMode2.ANAGLYPH:
    //   s = Math.min(dw / iw, dh / ih);
    //   h_port = 1;
    //   url = photo_.stereo;
    //   break;
    }
    $(".3d").remove();

    style = {
      position: "absolute",
      width: Math.floor(s * iw * h_port),
      height: Math.floor(s * ih),
      background: "url(\"" + url + "\")",
      "background-size": Math.floor(s * iw) + "px " + Math.floor(s * ih) + "px"
    };

    left_div_ = $("<div/>", {class: "3d"}).css(style);
    right_div_ = $("<div/>", {class: "3d"})
      .css(style)
      .css("background-position", "-" + Math.floor(s * iw / 2) + "px 0px");

    contents = $("<div/>").addClass("midi3d");

    switch (mode_) {
    case TT_DisplayMode2.PARALLEL:
      right_div_.css("left", Math.floor(s * iw / 2));
      right_div_.show();
      contents.css({width: 2 * (Math.floor(s * iw / 2)), height: Math.floor(s * ih)});
      break;

    case TT_DisplayMode2.CROSSEYE:
      left_div_.css("left", Math.floor(s * iw / 2));
      right_div_.show();
      contents.css({width: 2 * (Math.floor(s * iw / 2)), height: Math.floor(s * ih)});
      break;

    case TT_DisplayMode2.OVERLAY:
    case TT_DisplayMode2.ACTIVE:
      vis_w = Math.floor((iw * s) / 2);
      vis_h = Math.floor(ih * s);
      x_off = Math.floor(photo_.stereo_dx * iw * s);
      y_off = Math.floor(photo_.stereo_dy * iw * s);
      if (x_off <= 0) {
        left_div_.css("left", x_off);
        vis_w += x_off;
      } else {
        right_div_.css("left", -x_off);
        vis_w -= x_off;
      }
      if (y_off <= 0) {
        left_div_.css("top", y_off);
        vis_h += y_off;
      } else {
        right_div_.css("top", -y_off);
        vis_h -= y_off;
      }
      if (mode_ === TT_DisplayMode2.OVERLAY) {
	right_div_.css("opacity", 0.5);
	left_div_.css("opacity", 1);
      } else {
	right_div_.hide();
      }
      contents.css({width: vis_w, height: vis_h});
      break;

    case TT_DisplayMode2.LEFT:
    // case TT_DisplayMode2.ANAGLYPH:
      right_div_.hide();
      contents.css({width: left_div_.width(), height: left_div_.height()});
      break;
    }

    contents.keydown(key_down);
    contents.append(left_div_);
    contents.append(right_div_);
    div_.append(contents);
  }

  function render() {
    clearTimeout(ticker_);
    if (!cur_img_) {
      div_.html("Missing image.");
      return;
    }
    create_divs();
    if (mode_ === TT_DisplayMode2.ACTIVE) {
      ticker_ = setTimeout(function () { swap_sides(); }, mode_.delay);
    }
  }

  function clear() {
    clearTimeout(ticker_);
    if (left_div_ !== null) {
      left_div_.remove();
      left_div_ = null;
    }
    if (right_div_ !== null) {
      right_div_.remove();
      right_div_ = null;
    }
  }

  function set(div, photo, cur_img, mode, pw, ph) {
    div_= div;
    div_orig_width_ = pw;
    div_orig_height_ = ph;
    clear();

    if (!(photo.hasOwnProperty("stereo_dx") && photo.hasOwnProperty("stereo_dy"))) {
      return;
    }
    photo_ = photo;
    mode_ = mode;
    cur_img_ = cur_img;
    render();
  }

  function pixels_per_cm(cm) {
    var d, r;
    d = $("<div style='position:absolute; left:-100%; top:-100%; width:" + cm + "cm'></div>");
    $(document.body).append(d);
    r = d.width();
    d.remove();
    return r;
  }

  max_parallel_dw_ = 100;

  div_ = null;
  div_orig_width_ = 0;
  div_orig_height_ = 0;

  left_div_ = null;
  right_div_ = null;

  cur_img_ = null;
  photo_ = null;

  mode_ = TT_DisplayMode2.PARALLEL;
  ticker_ = null;

  return {
    mode: function () { return mode_; },
    control: function () { return mode_.control; },
    key_down: key_down,
    clear: clear,
    set: set,
    initialize: function () {
      max_parallel_dw_ = pixels_per_cm(7.0);
    }
  };
}());
