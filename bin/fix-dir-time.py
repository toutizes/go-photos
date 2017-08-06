#!/usr/bin/python
#
# Make sure that the mtime of directories are at least as high as the
# mtime of all the files and directories they contain.
#
# Usage fix-dir-time dir...

import os
import sys

assert len(sys.argv) > 1, 'Pass root dirs on the command line'

mtimes_cache = {}

def getmtime(path):
  if path not in mtimes_cache:
    mtimes_cache[path] = os.path.getmtime(path)
  return mtimes_cache[path]


def setmtime(path, time):
  os.utime(dir, (time, time))
  mtimes_cache[path] = time


for root_dir in sys.argv[1:]:
  for dir, subdirs, files in os.walk(root_dir, topdown=False):
    # make files and subdirs absolute paths
    files = [os.path.join(dir, f) for f in files]
    subdirs = [os.path.join(dir, f) for f in subdirs]
    # get the mtime of the directory and all its subdirs and files. 
    dir_time = getmtime(dir)
    # If there is anything in the directory check timestamps.
    if subdirs or files:
      sub_time = max([getmtime(x) for x in subdirs + files])
      # fix the directory time if needed.
      if sub_time > dir_time:
        print 'Touching: ' + dir
        setmtime(dir, sub_time)
