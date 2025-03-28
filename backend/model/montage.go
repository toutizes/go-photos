package model

import (
  "io/ioutil"
  "log"
  "net/http"
  "os/exec"
  "path"
  "strconv"
  "strings"
)

func montageSpec(spec string) (geometry string, ids []int) {
  splits := strings.Split(spec, "-")
  if len(splits) < 2 {
    return "", nil
  }
  ids = make([]int, len(splits) - 1)
  for i, s := range splits[1:] {
    v, err := strconv.Atoi(s)
    if err != nil {
      return "", nil
    }
    ids[i] = v
  }
  geometry = splits[0]
  return geometry, ids
}

func servePath(w http.ResponseWriter, path string) bool {
  bytes, err := ioutil.ReadFile(path)
  if err != nil {
    return false
  } else {
    w.Header().Set("Content-Type", "image/jpeg")
    w.Write(bytes)
    return true;
  }
}

func HandleMontage2(w http.ResponseWriter, r *http.Request, db *Database) {
  splits := strings.Split(r.URL.Path, "/")
  if len(splits) == 0 {
    return
  }
  spec := splits[len(splits) - 1]
  geo, ids := montageSpec(spec)
  mont := path.Join(db.MontagePath(), spec)
  if servePath(w, mont) {
    return
  }
  cmd := exec.Command(*BinRoot + "magick")
  args := make([]string, 0, len(ids) + 16)
  args = append(args, "montage")
  for _, id := range ids {
    args = append(args, db.MiniPath(id))
  }
  // First resize to fill the square while maintaining aspect ratio
  args = append(args, "-resize");
  args = append(args, geo + "^");
  // Center the image
  args = append(args, "-gravity");
  args = append(args, "center");
  // Crop to exact square size
  args = append(args, "-extent");
  args = append(args, geo);
  // Ensure each image is treated independently
  args = append(args, "+repage");
  // Create a single row montage
  args = append(args, "-tile");
  args = append(args, "x1");
  // No spacing between images
  args = append(args, "-geometry");
  args = append(args, geo + "+0+0");
  args = append(args, mont);
  cmd.Args = args;
  
  output, err := cmd.CombinedOutput()
  if err != nil {
    log.Printf("montage failed (%s): %s\n", strings.Join(args, " "), err)
    log.Printf("Full output: %s\n", string(output))
  }
  servePath(w, mont)
}
