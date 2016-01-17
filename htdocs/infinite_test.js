source = function() {
  return {
    items_per_page: function() { return 5; },
    make_item_div: function(index) {
      div = $("<div/>");
      div.append("<span>item " + index + "</span>");
      div.append($("<img/>", {height: 100, width: 100}));
      return div;
    }
  };
}();

function infinite_test() {
  infinite = TT_infinite($("#container"), $("#contents"), source);
  infinite.display(100);
}
