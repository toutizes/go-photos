/*global $, console, TT_Fetcher2, TT_Montage, TT_DB5, TT_Slider5, TT_Preloader*/
/*jslint browser:true, nomen: true, unparam: true*/
/* Uses jquery extensions: scrollintoview, waypoints. */
var N = 4;
var last = 0;
var first = -1;
var cur = null;
var horizontal = true;

function wp_last(d) {
  if (d == "down" || d == "right") {
    console.log("last: " + this.element.id + ": " + d);
    this.destroy();
    append(true);
  }
}

function append(add_wp) {
  var i, opt, div;
  for (i = 0; i < N; ++i) {
    div = $("<div/>").attr("id", "x" + last).css("display", "inline-block").html("last " + last);
    $("#contents").append(div);
    last += 1;
  }
  if (add_wp) {
    opt = {
      context: "#container",
      horizontal: true,
      offset: "right-in-view"
    };
    div.waypoint(wp_last, opt);
  }
  return div;
}

function wp_first(d) {
  if (d == "up" || d == "left") {
    console.log("first: " + this.element.id + ": " + d);
    this.destroy();
    prepend(true);
  }
}

function prepend(add_wp) {
  var i, opt, div, added_h;
  if (!horizontal) {
    console.log("top " + $("#container").scrollTop());
  } else {
    console.log("left " + $("#container").scrollLeft());
  }
  added_h = 0;
  for (i = 0; i < N; ++i) {
    div = $("<span/>").attr("id", "x" + first).html("first " + first);
    $("#contents").prepend(div);
    added_h += horizontal ? div.outerWidth() : div.outerHeight();
    first -= 1;
  }
  if (add_wp) {
    opt = {
      context: "#container",
      horizontal: true,
      offset: -div.outerWidth()
    };
    div.waypoint(wp_first, opt);
  }
  if (!horizontal) {
    $("#container").scrollTop($("#container").scrollTop() + added_h);
  } else {
    $("#container").scrollLeft($("#container").scrollLeft() + added_h);
  }
  console.log("prepended");
  return div;
}

function scroll_into_view(container, div) {
  if (!horizontal) {
    div_top = div.offset().top;
    div_bot = div_top + div.outerHeight();
    cont_top = container.offset().top;
    cont_bot = cont_top + container.height();
    if (div_top < cont_top) {
      container.scrollTop(container.scrollTop() - (cont_top - div_top));
    } else if (div_bot > cont_bot) {
      container.scrollTop(container.scrollTop() + (div_bot - cont_bot));
    }
  } else {
    div_left = Math.floor(div.offset().left);
    div_right = div_left + div.outerWidth();
    cont_left = Math.floor(container.offset().left);
    cont_right = cont_left + container.width();
    console.log(div_left + " " + div_right);
    console.log(cont_left + " " + cont_right);
    if (div_left < cont_left) {
      container.scrollLeft(container.scrollLeft() - (cont_left - div_left));
    } else if (div_right > cont_right) {
      container.scrollLeft(container.scrollLeft() + (div_right - cont_right));
    }
  }
}

function show(new_cur) {
  var div = null, div_top, div_bot, cont_top, cont_bot;
  if (cur !== new_cur) {
    if (cur !== null) {
      $("#x" + cur).css("background", "pink");
    }
    div = $("#x" + new_cur);
    if (div) {
      scroll_into_view($("#container"), div);
    }
    cur = new_cur;
    $("#x" + cur).css("background", "red");
  }
}

function key_down(event) {
  switch (event.which) {
  case 38:                        // up arrow
    event.preventDefault();
    show(cur - 1);
    console.log("shown");
    break;
  case 40:                        // down arrow
    event.preventDefault();
    while (cur >= last) {
      append(true);
    }
    show(cur + 1);
    break;
  default:
    show(cur);
    break;
  }
}

function initialize() {
  $(document.body).keydown(key_down);
  append(true);
  append(true);
  prepend(true);
  show(0);
}

$(document).ready(initialize);
