/*global $, console, TT_DisplayMode2*/
/*jslint browser:true, nomen: true, unparam: true*/
var TT_Slider5 = (function () {
  "use strict";

  var images_midi_,
    images_midi_container_,
    images_maxi_container_,
    midi_model_,
    sequence_,
    threed_,
    NONE = 0,
    SLIDE_LEFT = 1,
    SLIDE_RIGHT = 2,
    FULL = 3,
    mode_ = NONE,
    st_mode_ = TT_DisplayMode2.PARALLEL,
    IMAGE_PAD = 8,
    MIDI_PAD = 2;

  // Center the Image "img" and add it to "div".
  function center_image(div, photo, img, pw, ph) {
    var b, iw, ih, s;
    // This is the border-width of the image, get it from the image style.
    if (photo && photo.hasOwnProperty("stereo_dx")) {
      $("#controls3d").removeClass("hidden");
      threed_.set(div, photo, img, st_mode_, pw, ph);
    } else {
      $("#controls3d").addClass("hidden");
      b = 5;
      iw = img.naturalWidth + 2 * b + 2 * IMAGE_PAD;
      ih = img.naturalHeight + 2 * b + 2 * IMAGE_PAD;
      s = Math.min(pw / iw, ph / ih);
      div.append($(img)
		 .addClass("midiimg")
		 .css({width: Math.floor(img.naturalWidth * s),
		       height: Math.floor(img.naturalHeight * s)}));
    }
  }

  function recenter_midi(index, midi, pw, ph) {
    // Works for FULL too.
    if ($(midi).data() !== null) {
      $(midi).css({left: $(midi).data().index * (pw + 2 * MIDI_PAD),
                   width: pw, height: ph});
      center_image(midi, null, $(midi).children()[0], pw, ph);
    }
  }

  function resized() {
    var container = (mode_ === FULL) ? images_maxi_container_ : images_midi_container_,
      pw = container.width(),
      ph = container.height(),
      images_midi_data = images_midi_.data();
    // No animations while resizing.
    images_midi_.removeClass("slider");
    $.each(images_midi_.children(),
           function (index, midi) { recenter_midi(index, $(midi), pw, ph); });
    recenter_midi(0, $("#midi-full"), pw, ph);
    if (images_midi_data !== null) {
      images_midi_.css({left: -images_midi_data.index * (pw + 2 * MIDI_PAD) + MIDI_PAD});
    }
  }

  function img_ready(h, photo, image, slide_sequence, mode, cb) {
    var container = mode === FULL ? images_maxi_container_ : images_midi_container_,
      pw = container.width(),
      ph = container.height(),
      midi,
      midi_index;
    if (sequence_ !== slide_sequence) {
      return;
    }
    // Remove the old midi images, mark the remaining ones for removal.
    images_midi_.find("[to_remove=true]").remove();
    images_midi_.children().attr("to_remove", "true");

    // Always remove the full image.
    images_maxi_container_.find(".midi").remove();
    images_maxi_container_.addClass("hidden");

    mode_ = mode;
    switch(h.st_mode) {
    case TT_DisplayMode2.PARALLEL.code: st_mode_ = TT_DisplayMode2.PARALLEL; break;
    case TT_DisplayMode2.CROSSEYE.code: st_mode_ = TT_DisplayMode2.CROSSEYE; break;
    case TT_DisplayMode2.ACTIVE.code: st_mode_ = TT_DisplayMode2.ACTIVE; break;
    case TT_DisplayMode2.OVERLAY.code: st_mode_ = TT_DisplayMode2.OVERLAY; break;
    case TT_DisplayMode2.LEFT.code: st_mode_ = TT_DisplayMode2.LEFT; break;
    default: st_mode_ = null;
    }

    // Create the midi image to add.
    midi = midi_model_.clone(true).attr("id", "midi-" + sequence_);
    center_image(midi, photo, image, pw, ph);
    if (mode === NONE) {
      images_midi_.empty();
      images_midi_.append(midi);
      midi_index = 0;
    } else if (mode === SLIDE_LEFT) {
      midi_index = images_midi_.children().last().data().index + 1;
      images_midi_.append(midi);
    } else if (mode === SLIDE_RIGHT) {
      midi_index = images_midi_.children().first().data().index - 1;
      images_midi_.prepend(midi);
    } else if (mode === FULL) {
      images_midi_.empty();
      // Change id to "midi-full" so we can easily remove it later.
      midi.attr("id", "midi-full");
      images_maxi_container_.append(midi);
      images_maxi_container_.removeClass("hidden");
      midi_index = 0;
    }
    midi.data({index: midi_index});
    midi.css({left: midi_index * (pw + 2 * MIDI_PAD), top: 0, width: pw, height: ph});
    images_midi_.data({index: midi_index});
    images_midi_.css({left: -midi_index * (pw + 2 * MIDI_PAD) + MIDI_PAD});
    if (cb !== null) {
      cb();
    }
  }

  function display(h, photo, img_url, mode, cb) {
    var slide_sequence, image;
    // Use animations while sliding.
    images_midi_.addClass("slider");
    sequence_ += 1;
    slide_sequence = sequence_;
    $("#midi-full .midiimg").attr("src", "loading-noir.webp");
    image = new Image();
    image.onload = function () { img_ready(h, photo, image, slide_sequence, mode, cb); };
    image.src = img_url;
  }

  function clear() {
    images_midi_.empty();
    images_midi_.css({left: 0});
    images_maxi_container_.find(".midi").remove();
  }

  function center(h, photo, img_url, cb) {
    display(h, photo, img_url, NONE, cb);
  }

  function slide_left(h, photo, img_url, cb) {
    display(h, photo, img_url, SLIDE_LEFT, cb);
  }

  function slide_right(h, photo, img_url, cb) {
    display(h, photo, img_url, SLIDE_RIGHT, cb);
  }

  function show_full(h, photo, img_url, cb) {
    display(h, photo, img_url, FULL, cb);
  }

  function hide() {
    images_maxi_container_.addClass("hidden");
  }

  function show() {
    images_maxi_container_.removeClass("hidden");
  }

  function initialize(threed) {
    images_midi_container_ = $("#images-midi-container");
    images_maxi_container_ = $("#images-maxi-container");
    images_midi_ = $("#images-midi");
    midi_model_ = $("#midi-model");
    sequence_ = 0;
    threed_ = threed;
  }

  return {
    clear: clear,
    hide: hide,
    show: show,
    center: center,
    slide_left: slide_left,
    slide_right: slide_right,
    show_full: show_full,
    resized: resized,
    initialize: initialize
  };
}());
