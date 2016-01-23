/*global $, console, */
/*jslint browser:true, nomen: true*/
function tt_Infinite(container, contents, horizontal, source) {
  /* Uses jquery extensions: waypoints. */
  "use strict";
  
  // Div where to put the items to display.
  var contents_ = contents;

  // Scrolling container for 'contents_'.
  var container_ = container;

  // Infinite data source.  It provides:
  // - items_per_page(): The number of items that fit in a 'page'.
  // - make_item_div(index): Returns an element to display for index.
  //   It should return null if the requested index is out of bounds.
  var source_ = source;

  // True if horizontal, false if vertical;
  var horizontal_ = horizontal;
  
  var first_item_index_ = -1; // Index - 1 of the first item present in the container.
  var last_item_index_ = 0; // Index + 1 of the last item present in the container.
  var first_waypoint_ = null;	// Waypoint of first item.
  var last_waypoint_ = null;	// Waypoint of last item.

  // Make sure 'div' is visible, vertical case.
  function scroll_into_view_vert(div) {
    var div_top = Math.floor(div.offset().top);
    var div_bot = div_top + div.outerHeight();
    var cont_top = Math.floor(container_.offset().top);
    var cont_bot = cont_top + container_.height();
    if (div_top < cont_top) {
      container_.scrollTop(container_.scrollTop() - (cont_top - div_top));
    } else if (div_bot > cont_bot) {
      container_.scrollTop(container_.scrollTop() + (div_bot - cont_bot));
    }
  }

  // Make sure 'div' is visible, horizontal case.
  function scroll_into_view_hori(div) {
    var div_left = Math.floor(div.offset().left);
    var div_right = div_left + div.outerWidth();
    var cont_left = Math.floor(container_.offset().left);
    var cont_right = cont_left + container_.width();
    if (div_left < cont_left) {
      container_.scrollLeft(container_.scrollLeft() - (cont_left - div_left));
    } else if (div_right > cont_right) {
      container_.scrollLeft(container_.scrollLeft() + (div_right - cont_right));
    }
  }

  // Make sure the content for 'index' is visible.
  function scroll_into_view(index) {
    if (index < first_item_index_) {
      // add more and continue?
      return;
    } 
    if (index >= last_item_index_) {
      // add more and continue?
      return;
    }
    var children = contents_.children();
    var rel_index = index - first_item_index_ - 1;
    if (rel_index < 0 || rel_index >= children.length) {
      // Nothing to do.
      return;
    }
    var div = children.eq(rel_index);
    if (horizontal_) {
      scroll_into_view_hori(div);
    } else {
      scroll_into_view_vert(div);
    }
  }

  // Memorize what it needed to restore the position of contents_ later
  // after adding elements to it.
  function contents_pos() {
    var first = contents_.children().eq(0);
    return {first: first, top: first.offset().top, left: first.offset().left};
  }

  // Restore the position of contents_.  'before' is what
  // contents_pos() returned before adding elements to contents_.
  function restore_contents_pos(before) {
    var change;
    if (horizontal_) {
      change = before.first.offset().left - before.left;
      container_.scrollLeft(container_.scrollLeft() + change);
    } else {
      change = before.first.offset().top - before.top;
      container_.scrollTop(container_.scrollTop() + change);
    }
  }

  // Callback for waypoints at the end of the displayed items.
  function wp_after(d) {
    if (d === "down" || d === "right") {
      this.disable();		// 'this' is the waypoint.
      add_next_page_of_items();
    }
  }

  // Callback for waypoints at the beginning of the displayed items.
  function wp_before(d) {
    if (d === "up" || d === "left") {
      this.disable();		// 'this' is the waypoint.
      add_prev_page_of_items();
    }
  }

  // Add a waypoint callback for "div".
  function add_waypoint(div, after) {
    var waypoint_options = {
      context: container_,
      horizontal: horizontal_
    };
    var callback;
    if (after) {
      waypoint_options.offset = horizontal_ ? "right-in-view" : "bottom-in-view";
      callback = wp_after;
    } else {
      waypoint_options.offset = horizontal_ ? -div.outerWidth() : -div.outerHeight();
      callback = wp_before;
    }
    return div.waypoint(callback , waypoint_options)[0];
  }

  // Destroy a waypoint, return null.
  // Usage: xx_ = destroy_waypoint(wxx_);
  function destroy_waypoint(wp) {
    if (wp !== null) {
      wp.destroy();
    }
    return null;
  }

  // Append the next page of items.  If there are more item to add,
  // add a Waypoint to the last added item to add the next page.
  function add_next_page_of_items() {
    var i, item_div;
    var n = source_.items_per_page();
    last_waypoint_ = destroy_waypoint(last_waypoint_);
    for (i = 0; i < n; i++) {
      item_div = source_.make_item_div(last_item_index_);
      if (item_div === null) {
	// No more minis to add.  No waypoints to set either.
	break;
      }
      contents_.append(item_div);
      last_item_index_ += 1;
    }
    if (item_div !== null) {
      last_waypoint_ = add_waypoint(item_div, true /* after */);
    }
  }

  // Same as add_next_page_of_items() but adds the previous page.
  // Takes special care to not change the scroll position of the
  // container.
  function add_prev_page_of_items() {
    var n = source_.items_per_page();
    var pos_before = contents_pos();
    first_waypoint_ = destroy_waypoint(first_waypoint_);
    var i, item_div;
    for (i = 0; i < n; i++) {
      item_div = source_.make_item_div(first_item_index_);
      if (item_div === null) {
	// No more items to add.  No waypoints to set either.
	break;
      }
      contents_.prepend(item_div);
      first_item_index_ -= 1;
    }
    restore_contents_pos(pos_before);
    if (item_div !== null) {
      // Add a waypoint if we need it.
      first_waypoint_ = add_waypoint(item_div, false /* after */);
    }
  }

  // Same as add_next_page_of_items() but add a page centered
  // around the passed in index.
  function add_center_page_of_items(index) {
    // Center the indices where to add images.
    first_item_index_ = index - 1;
    last_item_index_ = index;
    var first_div = null;
    var last_div = null;
    var cur_div = null;
    // Add a full page before the current image.
    var n = source_.items_per_page();
    var i;
    for (i = 0; i < n; i++) {
      first_div = source_.make_item_div(first_item_index_);
      if (first_div === null) {
	// No more minis to add.
	break;
      }
      contents_.prepend(first_div);
      first_item_index_ -= 1;
    }
    // Add a full page after the current image.
    for (i = 0; i < n; i++) {
      last_div = source_.make_item_div(last_item_index_);
      if (last_div === null) {
	// No more minis to add.
	break;
      }
      if (cur_div === null) {
	cur_div = last_div;
      }
      contents_.append(last_div);
      last_item_index_ += 1;
    }
    scroll_into_view(index);
    // Add waypoints to get more data.
    if (first_div) {
      first_waypoint_ = add_waypoint(first_div, false /* bottom */);
    }
    if (last_div) {
      last_waypoint_ = add_waypoint(last_div, true /* bottom */);
    }
  }

  // Rebuild the contents from scratch around 'index'.
  function display(index) {
    contents_.empty();
    add_center_page_of_items(index);
  }

  // Clear the elements, destroy the waypoints.
  function destroy() {
    contents_.empty();
    first_waypoint_ = destroy_waypoint(first_waypoint_);
    last_waypoint_ = destroy_waypoint(last_waypoint_);
  }

  return {
    display: display,
    scroll_into_view: scroll_into_view,
    horizontal: function() { return horizontal_; },
    destroy: destroy
  };
}
