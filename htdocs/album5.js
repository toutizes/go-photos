/*global $, console, TT_Fetcher2, TT_Montage, TT_DB5, tt_Infinite*/
/*jslint browser:true, nomen: true*/
var TT_Album5 = (function () {
  "use strict";

  var album_container_ = null;	// Container for albums.
  var model_ = null;		// Model to clone for albums.
  var montage_ = null;		// Montage objects for album images.
  var infinite_ = null;		// Infinite scroller object.
  var albums_ = null;		// List of albums to display.

  function req_album() {
    var q = "album:" + $(this).data().album.id;
    if (q.indexOf(" ") !== -1) {
      q = "\"" + q + "\"";
    }
    TT_DB5.req_string(q, null);
  }

  function make_album_div(index) {
    if (index < 0 || index >= albums_.length) {
      return null;
    }
    var album = albums_[index];
    var div = model_.clone().attr("id", "album-" + index);
    div.css(montage_.bg_style(index)).data({album: album}).click(req_album);
    var desc = div.find(".album-desc");
    desc.find(".album-title").append(album.title);
    var num_photos_txt = album.numPhotos + " photo";
    if (album.numPhotos !== 1) {
      num_photos_txt += "s";
    }
    desc.find(".album-num-photos").append(num_photos_txt);
    return div;
  }

  function albums_per_page() {
    var h_albums = album_container_.width() / 150;
    var p_albums = h_albums * (album_container_.height() / 150);
    var n = Math.min(p_albums, albums_.length);
    return Math.floor(n + p_albums);
  }

  function by_date_updated(a, b) {
    return b.updated - a.updated;
  }

  function albums_ready(albums) {
    albums_ = albums.sort(by_date_updated);
    var ids = $.map(albums_, function (a) { return a.coverId; });
    montage_ = TT_Montage.create(8, Math.floor(model_.width()), ids);
    infinite_.display(0);
  }

  function display() {
    TT_Fetcher2.getAllAlbums(albums_ready);
  }

  function hide() {
    album_container_.addClass("hidden");
  }

  function show() {
    album_container_.removeClass("hidden");
  }

  function initialize() {
    album_container_ = $("#album-container");
    model_ = $("#album-model");
    infinite_ = tt_Infinite(album_container_, $("#album-contents"),
			    false /* horizontal */,
    			    { items_per_page: albums_per_page,
    			      make_item_div: make_album_div
    			    });
  }

  return {
    display: display,
    resized: function () { return null; },
    hide: hide,
    show: show,
    initialize: initialize
  };
}());
