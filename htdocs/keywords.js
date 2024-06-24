/*global $, console*/
/*jslint browser:true, nomen: true, unparam: true*/
var TT_Keywords = (function () {
  "use strict";

  var container_ = null;	// Container for all keywords.
  var model_ = null;		// Model for keywords.
  var album_ = null;		// Container for album name.
  var keywords_ = null;		// Container for image keywords.

  function req_keyword(e) {
    var data = $(this).data();
    TT_DB5.req_string(data.kw, data.image);
  }

  function add_keyword(keyword, keyword_q, image_id, container) {
    var kw = model_.clone().attr("id", null).text(keyword);
    if (keyword_q.indexOf(" ") !== -1) {
      keyword_q = "\"" + keyword_q + "\"";
    }
    kw.data({kw: keyword_q, image: image_id});
    kw.click(req_keyword);
    container.append(kw);
    container.append(" ");
  }

  function display(image) {
    album_.empty();
    keywords_.empty();
    const date_str = new Date(image.timestamp * 1000).toLocaleDateString('en-CA');
    add_keyword(date_str, date_str, image.id, album_);
    add_keyword("Tout l'album", "album:" + image.albumId, image.id, album_);
    var i, keyword;
    for (i = 0; i < image.keywords.length; i++) {
      keyword = image.keywords[i];
      add_keyword(keyword, keyword, image.id, keywords_);
    }
  }

  function hide() {
    container_.addClass("hidden");
  }

  function show() {
    container_.removeClass("hidden");
  }

  function initialize(slider) {
    container_ = $("#keywords-container");
    model_ = $("#keywords-model");
    keywords_ = $("#keywords");
    album_ = $("#keywords-album");
  }

  return {
    display: display,
    hide: hide,
    show: show,
    initialize: initialize
  };
}());
