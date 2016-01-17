/*global $,console*/
/*jslint browser: true, unparam: true, vars: true, nomen: true*/

// Ajax interface to toutizes.com Go server..
var TT_Fetcher2 = (function () {
  "use strict";

  var all_albums_ = null, interned_strings_ = {};

  function ajax(url, params, cb) {
    var ajax_params = {
      url: url,
      type: "POST",
      dataType: "json",
      data: params,
      success: cb,
      error: function (jqXHR, textStatus, errorThrown) {
        console.log("error " + textStatus);
        console.log(errorThrown);
      }
    };
    // console.log(ajax_params);
    $.ajax(ajax_params);
  }

  function makeAlbum(jalb) {
    return {
      id: jalb.Id,
      timestamp: jalb.Ats,
      updated: jalb.Dts,
      title: jalb.Id,
      numPhotos: jalb.Nimgs,
      coverId: jalb.Cov
    };
  }

  function gotAlbumFeed(jalbs, cb) {
    var albums = [], jalb, i;
    for (i = 0; i < jalbs.length; i++) {
      jalb = jalbs[i];
      if (jalb.Nimgs > 0) {
        albums.push(makeAlbum(jalb));
      }
    }
    all_albums_ = albums;
    cb(all_albums_);
  }

  function getAllAlbums(cb) {
    if (all_albums_ !== null) {
      cb(all_albums_);
    } else {
      ajax("q",
           {"q": "albums:", "kind": "album"},
           function (feed) { gotAlbumFeed(feed, cb); });
    }
  }

  function albumIdOrTitle(x) {
    return x;
  }

  function makeKeywords(kwds) {
    var keywords = [], i, kw, kwi;
    for (i = 0; i < kwds.length; i++) {
      kw = kwds[i];
      kwi = interned_strings_[kw];
      if (kwi === undefined) {
        interned_strings_[kw] = kw;
        kwi = kw;
      }
      keywords.push(kwi);
    }
    return keywords;
  }

  function makePhoto(jphoto) {
    var photo;
    // console.log(jphoto);
    photo = {
      id: jphoto.Id,
      albumId: jphoto.Ad,
      filename: jphoto.In,
      timestamp: jphoto.Its,
      updated: jphoto.Fts,
      keywords: makeKeywords(jphoto.Kwd),
      w: jphoto.W,
      h: jphoto.H
    };
    if (jphoto.Stereo) {
      photo.stereo_dx = jphoto.Stereo.Dx;
      photo.stereo_dy = jphoto.Stereo.Dy;
    }
    return photo;
  }

  function makeNoResultPhoto() {
    return {
      id: 0,
      albumId: "rien",
      filename: "rien-a-voir.jpg",
      timestamp: 0,
      updated: 0,
      keywords: []
    };
  }

  function by_timestamp_decreasing(a, b) {
    return b.timestamp - a.timestamp;
  }

  function by_timestamp_increasing(a, b) {
    return a.timestamp - b.timestamp;
  }

  function gotPhotoFeed(query, jphotos, cb) {
    // console.log("Got photo feed: ", jphotos);
    var photos = [], i;
    if (jphotos.length === 0) {
      photos.push(makeNoResultPhoto());
    } else {
      for (i = 0; i < jphotos.length; i++) {
        photos.push(makePhoto(jphotos[i]));
      }
    }
    // console.log("Got photos: ", photos);
    if (query.match(/^album:/)) {
      photos = photos.sort(by_timestamp_increasing);
    } else {
      photos = photos.sort(by_timestamp_decreasing);
    }
    cb(photos);
  }

  function queryPhotos(query, cb) {
    ajax("q",
         {"q": query, "kind": "photo"},
         function (feed) { gotPhotoFeed(query, feed, cb); });
  }

  function setStereo(id, dx, dy) {
    ajax("set",
         {"id": id, "dx": dx, "dy": dy},
         function (feed) { console.log(feed); });
  }

  function clearStereo(id) {
    ajax("set",
         {"id": id},
         function (feed) { console.log(feed); });
  }

  // Public interface.
  return {
    queryPhotos: queryPhotos,
    getAllAlbums: getAllAlbums,
    albumIdOrTitle: albumIdOrTitle,
    setStereo: setStereo,
    clearStereo: clearStereo
  };
}());
