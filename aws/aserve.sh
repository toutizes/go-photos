#!/bin/bash
LOG_FILE="/mnt/photos/daemon/aserve.log.$(date '+%Y-%m-%d:%H:%m:%S')"

/mnt/photos/bin/aserve \
  --bin_root=/usr/bin/ \
  --db_root=/mnt/photos \
  --static_root=/mnt/photos/htdocs \
  --port=8080 \
  --num_cpu=2 \
  --minifier_threads=2 \
  "$@" \
  &> "$LOG_FILE" &

# process id of the launched aserve
echo $!
