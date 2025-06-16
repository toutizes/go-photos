import os
import sys

LR="/Users/matthieu/Pictures/Photos"
GD="/Users/matthieu/Google Drive/My Drive/Photos"


def list_files_rec(directory):
    """Recursively lists all files in the given directory.

    Returns a list of file paths as strings, dropping the file extension.
    """
    files = []
    for filename in os.listdir(directory):
        full_path = os.path.join(directory, filename)
        if os.path.isfile(full_path):
            files.append(os.path.splitext(full_path)[0])
        elif os.path.isdir(full_path):
            files.extend(list_files_rec(full_path))
    return sorted(files)


def list_files(root, sub):
  return [f[len(root)+1:] for f in list_files_rec(os.path.join(root, sub))]


def compare(sub):
    """Compares two lists of files and returns the differences."""
    lr_set = set(list_files(LR, sub))
    gd_set = set(list_files(GD, sub))
    
    only_in_lr = lr_set - gd_set
    only_in_gd = gd_set - lr_set
    
    return {
        "lr": lr_set,
        "gd": gd_set,
        "only_lr": sorted(only_in_lr),
        "only_gd": sorted(only_in_gd),
        "both": sorted(lr_set & gd_set)
    }


comp = compare(sys.argv[1])

print("Files in LR: %d" % len(comp["lr"]))
print("Files in GD: %d" % len(comp["gd"]))
print("Files in both LR and GD: %d" % len(comp["both"]))

only_lr = comp["only_lr"]
if only_lr:
    print("Files only in LR: %d" % len(only_lr))
    for f in only_lr:
        print("  ", f)

