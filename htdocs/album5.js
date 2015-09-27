/*global $, console, TT_Fetcher2, TT_Montage, TT_DB5*/
/*jslint browser:true, nomen: true*/
var TT_Album5 = (function () {
  "use strict";

  var album_container_, albums_contents_, model_, image_size_, montage_,
    h_albums_ = null, h_albums_more_ = 0;

  function req_album() {
    var q = "album:" + $(this).data().album.id;
    if (q.indexOf(" ") !== -1) {
      q = "\"" + q + "\"";
    }
    TT_DB5.req_string(q, null);
  }

  function add_album(index, album, offset) {
    var div, desc, num_photos_txt, abs_index = offset + index;
    div = model_.clone().attr("id", "album-" + abs_index);
    div.css(montage_.bg_style(abs_index)).data({album: album}).click(req_album);
    desc = div.find(".album-desc");
    desc.find(".album-title").append(album.title);
    num_photos_txt = album.numPhotos + " photo";
    if (album.numPhotos !== 1) {
      num_photos_txt += "s";
    }
    desc.find(".album-num-photos").append(num_photos_txt);
    // desc.find(".album-keywords").append("Lorem ipsum dolor sit amet");
    albums_contents_.append(div);
  }

  function add_more_albums() {
    var h_albums = Math.floor(album_container_.width() / 150),
      N = h_albums_more_ + h_albums * 2;
    $.each(h_albums_.slice(h_albums_more_, h_albums_more_ + N),
           function (index, album) { add_album(index, album, h_albums_more_); });
    h_albums_more_ = h_albums_more_ + N;
    album_container_.mCustomScrollbar("update");
  }

  function show_albums(albums) {
    var h_albums = album_container_.width() / 150,
      p_albums = h_albums * (album_container_.height() / 150),
      N = Math.min(p_albums, albums.length);
    N = Math.floor(N + p_albums);
    albums_contents_.empty();
    $.each(albums.slice(0, N),
           function (index, album) { add_album(index, album, 0); });
    h_albums_ = albums;
    h_albums_more_ = N;
    album_container_.mCustomScrollbar("update");
  }

  function by_date_updated(a, b) {
    return b.updated - a.updated;
  }

  function albums_ready(albums) {
    var names;
    albums = albums.sort(by_date_updated);
    names = $.map(albums, function (a) { return a.coverId; });
    montage_ = TT_Montage.create(8, Math.floor(image_size_), names);
    show_albums(albums);
  }

  function display() {
    // var albums = h.mode === 0 ? NEWS_ALBUMS : 100;
    var albums = null;          // Real want all albums.
    TT_Fetcher2.getAllAlbums(albums_ready);
  }

  function hide() {
    album_container_.addClass("hidden");
  }

  function show() {
    album_container_.removeClass("hidden");
    album_container_.mCustomScrollbar("update");
  }

  function initialize() {
    var sb = {scrollInertia: 0,
              scrollButtons: {enable: true},
              callbacks: { onTotalScroll: add_more_albums,
                           onTotalScrollOffset: 300 }};
    album_container_ = $("#album-container");
    album_container_.mCustomScrollbar(sb);
    albums_contents_ = $("#album-container .mCSB_container");
    model_ = $("#album-model");
    image_size_ = model_.width();
  }

  return {
    display: display,
    add_more_albums: add_more_albums,
    resized: function () { return null; },
    hide: hide,
    show: show,
    initialize: initialize
  };
}());
