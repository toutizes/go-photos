#!/bin/bash -eu
# Installs a test image db in /tmp/aserve/db

# For debug info
#../waypoints/lib/jquery.waypoints.js
# ../waypoints/lib/waypoints.debug.js

JS_FILES="
../waypoints/lib/jquery.waypoints.min.js
../jquery-hashchange/jquery.ba-hashchange.min.js
../jquery-loading/dist/jquery.loading.min.js
../1130507/base64.js
htdocs/3d2.js
htdocs/preloader.js
htdocs/fetcher2.js
htdocs/hash.js
htdocs/album5.js
htdocs/image5.js
htdocs/iflow.js
htdocs/infinite.js
htdocs/keywords.js
htdocs/slider5.js
htdocs/db5.js
htdocs/montage.js
htdocs/util.js
"
cat $JS_FILES > /tmp/tt_db5.js

INSTALL_FILES="
htdocs/pic.html
htdocs/pic5.css
htdocs/icons/prev2.png
htdocs/icons/next2.png
htdocs/icons/active.png
htdocs/icons/parallel.png
htdocs/icons/crosseye.png
../malihu-custom-scrollbar-plugin/mCSB_buttons.png
/tmp/tt_db5.js
"

# Not needed anymore?
# ../jquery-loading/dist/jquery.loading.min.css

ROOT=/tmp/aserve/db
PHOTOS=$HOME/projects/test-photos

rm -rf $ROOT/htdocs
mkdir -p $ROOT/htdocs/db
cp $INSTALL_FILES $ROOT/htdocs/db

# mkdir $ROOT/mini
# mkdir $ROOT/midi
# ln -s $PHOTOS $ROOT/maxi
# ln -s $PHOTOS $ROOT/originals
