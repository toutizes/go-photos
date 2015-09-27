/*global $, TT_Hash, TT_Fetcher*/
/*jslint browser: true, unparam: true, plusplus: true, regexp: true*/
var TT_VID = (function () {
  "use strict";
  var videos, query;

  function padding(elt) {
    return elt.outerHeight(true) - elt.height();
  }

  function font_size(word, count, min, max) {
    var result;
    if (max === min) {
      return "100%";
    }
    result = 75 * (1 + Math.pow(count - min, 0.25) / Math.pow(max - min, 0.25));
    return Math.floor(result) + "%";
  }

  function size() {
    var results, available, height;
    results = $("#results");
    available = $(window).height() - padding($("body")) - results.offset().top;
    height = Math.max(available - padding(results), 100);
    if (height !== results.height()) {
      results.height(height);
    }
  }

  function search_q(e) {
    TT_Hash.set({q: $("#q").val()}, true);
  }

  function key_q(e) {
    if (e.which === 13) {
      search_q();
    }
  }

  function key_doc(e) {
    var player;
    if (!$.browser.msie) {
      if (e.target.nodeName !== "INPUT" && e.which === 32) {
        e.preventDefault();
        player = $("#player").find("#player_video")[0];
        if (player.paused) {
          player.play();
        } else {
          player.pause();
        }
      }
    }
  }

  function search_kw(kw) {
    if (kw.indexOf(" ") !== -1 || kw.indexOf("'") !== -1) {
      kw = "\"" + kw + "\"";
    }
    TT_Hash.set({q: kw}, true);
  }

  function word_cloud() {
    var cloud, i, keywords, j, kw, c;
    cloud = {};
    for (i = 0; i < videos.length; i++) {
      keywords = videos[i].keywords;
      for (j = 0; j < keywords.length; j++) {
        kw = keywords[j];
        if (kw !== query) {
          c = cloud[kw];
          if (c) {
            cloud[kw] = c + 1;
          } else {
            cloud[kw] = 1;
          }
        }
      }
    }
    return cloud;
  }

  function appendCloud(cloud, elt) {
    var keyword_model, max, min, c, index, f, kw, kw_elt;
    keyword_model = $("#keyword_model");
    max = 0;
    min = 100000;
    for (kw in cloud) {
      c = cloud[kw];
      max = Math.max(max, c);
      min = Math.min(min, c);
    }
    elt.empty();
    index = 0;
    f = function (e) { search_kw(kw); };
    for (kw in cloud) {
      kw_elt = keyword_model.clone();
      kw_elt.text(kw);
      kw_elt.click(f);
      kw_elt.css({"font-size": font_size(kw, cloud[kw], min, max)});
      elt.append(kw_elt);
      elt.append(" ");
      index += 1;
      if (index >= 60) {
        elt.append("... ");
        break;
      }
    }
  }

  function video_for_id(id) {
    var i, video;
    if (videos) {
      for (i = 0; i < videos.length; i++) {
        video = videos[i];
        if (video.id === id) {
          return video;
        }
      }
      return videos[0];
    }
    return null;
  }

  function appendKeywords(video, elt) {
    var keyword_model, keywords, i, f, kw, kw_elt;
    keyword_model = $("#keyword_model");
    keywords = video.keywords;
    elt.empty();
    f = function (e) {
      e.stopPropagation();
      search_kw(e.data.kw);
    };
    for (i = 0; i < keywords.length; i++) {
      if (i > 0) {
        elt.append(" ");
      }
      kw = keywords[i];
      kw_elt = keyword_model.clone();
      kw_elt.text(keywords[i]);
      kw_elt.click({kw: kw}, f);
      kw_elt.css({"font-size": font_size(kw, 10, 10, 10)});
      elt.append(kw_elt);
    }
  }

  function showVideo(video, play) {
    var holder, ie_loses, player;
    if (video) {
      holder = $("#player_holder");
      $(holder).find("#player_video").remove();
      holder.empty();
      if ($.browser.msie) {
        ie_loses = $("#ie_loses").clone();
        holder.prepend(ie_loses);
        ie_loses.show();
      } else {
        player = $("#player_model").clone();
        player.attr({id: "player"});
        $(player).find("#mp4").attr({src: video.mp4});
        $(player).find("#webm").attr({src: video.webm});
        holder.prepend(player);
        player.show();
        if (play) {
          $(player).find("#player_video")[0].play();
        }
      }
    }
  }

  function showResults(id) {
    var results, video_to_show, result_to_show, f, i, video, result, strip, scroll;
    results = $("#results");
    results.empty();
    results.append("Videos: " + videos.length);
    if (videos.length > 0) {
      video_to_show = videos[0];
      result_to_show = null;
      f = function (e) {
        TT_Hash.set({q: $("#q").val(), id: e.data.id}, false);
      };
      for (i = 0; i < videos.length; i++) {
        video = videos[i];
        result = $("#result_model").clone();
        result.attr({id: "result-" + i});
        appendKeywords(video, result.find("#keywords"));
        strip = $(result).find("#strip");
        strip.attr({src: video.mini});
        result.click({id: video.id}, f);
        results.append(result);
        if (video.id === id) {
          video_to_show = video;
          result_to_show = result;
        }
      }
      results.children().show();
      if (result_to_show) {
        scroll = result_to_show.position().top - results.position().top;
        results.scrollTop(scroll);
      }
      showVideo(video_to_show, false);
    }
  }

  function gotVideos(vs, h) {
    console.log("got " + vs.length);
    videos = vs;
    query = h.q;
    $("#q").val(query);
    appendCloud(word_cloud(), $("#player_keywords"));
    showResults(h.id);
  }

  function hashchange() {
    var h;
    h = TT_Hash.get();
    if (h.q === undefined) {
      h.q = "julien coline";
    }
    if (h.id === undefined) {
      h.id = null;
    }
    if (h.q !== query) {
      console.log(h.q);
      TT_Fetcher.queryVideos(h.q, function (vs) { gotVideos(vs, h); });
    } else if (h.id) {
      showVideo(video_for_id(h.id), true);
    } else if (videos.length > 0) {
      showVideo(videos[0], false);
    }
  }

  function initialize() {
    videos = [];
    query = null;

    $.loading({onAjax: true, delay: 300});
    $("#search").click(search_q);
    $("#q").keypress(key_q);
    $(document).keypress(key_doc);
    $(window).resize(size);
    $(window).hashchange(hashchange);
    size();
    hashchange();
  }

  function init() {
    $(document).ready(initialize);
  }

  return {
    init: init
  };
}());
