#!/bin/bash -eu
# Installs a test image db in /tmp/db

JS_FILES="
htdocs/custom-scrollbar-plugin/jquery.mCustomScrollbar.concat.min.js
htdocs/jquery.ba-hashchange.js
htdocs/jquery.loading.1.6.4.js
htdocs/3d2.js
htdocs/preloader.js
htdocs/fetcher2.js
htdocs/hash.js
htdocs/album5.js
htdocs/image5.js
htdocs/slider5.js
htdocs/db5.js
htdocs/base64.js
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
htdocs/custom-scrollbar-plugin/myCSB.css
htdocs/custom-scrollbar-plugin/mCSB_buttons.png
/tmp/tt_db5.js
"

ROOT=/tmp/db
PHOTOS=$HOME/projects/test-photos

rm -rf $ROOT/htdocs
mkdir -p $ROOT/htdocs/db
cp $INSTALL_FILES $ROOT/htdocs/db

# mkdir $ROOT/mini
# mkdir $ROOT/midi
ln -s $PHOTOS $ROOT/maxi
ln -s $PHOTOS $ROOT/originals
