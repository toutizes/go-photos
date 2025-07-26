package model

import (
  "encoding/json"
  "errors"
  "fmt"
  "net/http"
  "strconv"
  "log"
)

type ImageResults struct {
  Images []JsonImage
}

type DirectoryResults struct {
  Directories []JsonDirectory
}

type StringResults struct {
  Message string
}

// ImageInfo represents simplified image information for JSON responses
type ImageInfo struct {
  Id   int    `json:"id"`
  Name string `json:"name"`
}

// KeywordResponse represents a simplified keyword with image pairs for API responses
type KeywordResponse struct {
  Keyword      string      `json:"keyword"`
  Count        int         `json:"count"`
  RecentImages []ImageInfo `json:"recent_images"`
}

type KeywordResults struct {
  Keywords []KeywordResponse `json:"keywords"`
}

// KeywordGroupResponse represents a group of keywords that share the same images
type KeywordGroupResponse struct {
  Keywords     []KeywordResponse `json:"keywords"`
  RecentImages []ImageInfo       `json:"recent_images"`
  TotalWeight  float64           `json:"total_weight"`
  TotalCount   int               `json:"total_count"`
}

type KeywordGroupResults struct {
  Groups []KeywordGroupResponse `json:"groups"`
}

func queryImages(q string, db *Database) []*Image {
  if len(q) > 0 {
    qry := ParseQuery(q, db)
    imgs := make([]*Image, 0)
    for img := range qry {
      imgs = append(imgs, img)
    }
    return imgs
  } else {
    return nil
  }
}

func returnImages(w http.ResponseWriter, imgs []*Image) {
  enc := json.NewEncoder(w)
  res := make([]JsonImage, len(imgs))
  for i, img := range imgs {
    img.Json(&res[i])
  }
  enc.Encode(&res)
}

func returnDirectories(w http.ResponseWriter, imgs []*Image) {
  dirs := make(map[*Directory]bool)
  for _, img := range imgs {
    dirs[img.Directory()] = true
  }
  enc := json.NewEncoder(w)
  res := make([]JsonDirectory, len(dirs))
  i := 0
  for dir, _ := range dirs {
    dir.Json(&res[i])
    i += 1
  }
  enc.Encode(&res)
}

func HandleQuery(w http.ResponseWriter, r *http.Request, db *Database) {
  q := r.FormValue("q")
  kind := r.FormValue("kind")
  
  // Get user email from context
  userEmail := r.Context().Value("userEmail").(string)
  log.Printf("Query from %s: %q (kind: %s)", userEmail, q, kind)
  
  imgs := queryImages(q, db)
  switch {
  case kind == "album":
    returnDirectories(w, imgs)
  default:
    returnImages(w, imgs)
  }
}

func parseInt(r *http.Request, s string, err error) (int, bool, error) {
  if err != nil {
    return 0, false, err
  }
  vals, ok := r.Form[s]
  if !ok {
    return 0, false, nil
  }
  i, e := strconv.ParseInt(vals[0], 10, 32)
  return int(i), true, e
}

func parseFloat(r *http.Request, s string, err error) (float32, bool, error) {
  if err != nil {
    return 0.0, false, err
  }
  vals, ok := r.Form[s]
  if !ok {
    return 0, false, nil
  }
  f, e := strconv.ParseFloat(vals[0], 32)
  return float32(f), true, e
}

func HandleSet(w http.ResponseWriter, r *http.Request, db *Database) {
  var res StringResults
  err := r.ParseForm()
  id, has_id, err := parseInt(r, "id", err)
  dx, has_dx, err := parseFloat(r, "dx", err)
  dy, has_dy, err := parseFloat(r, "dy", err)
  var image *Image
  if err == nil {
    if !has_id {
      err = errors.New("Missing image id")
    }
  }
  if err == nil {
    image = db.Indexer().Image(id)
    if image == nil {
      err = errors.New(fmt.Sprintf("Unknown image id: %d", id))
    }
  }
  if err == nil {
    if has_dx && has_dy {
      // add and update the stereo info.
      if image.stereo == nil {
        image.stereo = new(Stereo)
      }
      image.stereo.Dx = dx
      image.stereo.Dy = dy
    } else if !has_dx && !has_dy {
      // delete the stereo info.
      image.stereo = nil
    } else {
      err = errors.New("Pass both dx and dy or none of them")
    }
  }
  if err == nil {
    err = db.SaveDirectory(image.Directory())
  }
  if err == nil {
    res.Message = "ok"
  } else {
    res.Message = err.Error()
  }
  enc := json.NewEncoder(w)
  enc.Encode(&res)
}

func HandleRecentKeywords(w http.ResponseWriter, r *http.Request, db *Database) {
  // Get user email from context
  userEmail := r.Context().Value("userEmail").(string)
  log.Printf("Recent keywords request from %s", userEmail)
  
  keywords := db.GetRecentActiveKeywords()
  
  // Convert from internal KeywordCount (with *Image) to external KeywordResponse (with simplified ImageInfo)
  responseKeywords := make([]KeywordResponse, len(keywords))
  for i, kw := range keywords {
    // Convert *Image to simplified ImageInfo
    simplifiedImages := make([]ImageInfo, len(kw.RecentImages))
    for j, img := range kw.RecentImages {
      simplifiedImages[j] = ImageInfo{
        Id:   img.Id,
        Name: img.Name(),
      }
    }
    
    responseKeywords[i] = KeywordResponse{
      Keyword:      kw.Keyword,
      Count:        kw.Count,
      RecentImages: simplifiedImages,
    }
  }
  
  enc := json.NewEncoder(w)
  result := KeywordResults{Keywords: responseKeywords}
  enc.Encode(&result)
}

func HandleRecentKeywordGroups(w http.ResponseWriter, r *http.Request, db *Database) {
  // Get user email from context
  userEmail := r.Context().Value("userEmail").(string)
  log.Printf("Recent keyword groups request from %s", userEmail)
  
  groups := db.GetRecentActiveKeywordGroups()
  
  // Convert from internal KeywordGroup to external KeywordGroupResponse
  responseGroups := make([]KeywordGroupResponse, len(groups))
  for i, group := range groups {
    // Convert keywords within the group
    responseKeywords := make([]KeywordResponse, len(group.Keywords))
    for j, kw := range group.Keywords {
      // Convert *Image to simplified ImageInfo
      simplifiedImages := make([]ImageInfo, len(kw.RecentImages))
      for k, img := range kw.RecentImages {
        simplifiedImages[k] = ImageInfo{
          Id:   img.Id,
          Name: img.Name(),
        }
      }
      
      responseKeywords[j] = KeywordResponse{
        Keyword:      kw.Keyword,
        Count:        kw.Count,
        RecentImages: simplifiedImages,
      }
    }
    
    // Convert group's recent images
    groupImages := make([]ImageInfo, len(group.RecentImages))
    for j, img := range group.RecentImages {
      groupImages[j] = ImageInfo{
        Id:   img.Id,
        Name: img.Name(),
      }
    }
    
    responseGroups[i] = KeywordGroupResponse{
      Keywords:     responseKeywords,
      RecentImages: groupImages,
      TotalWeight:  group.TotalWeight,
      TotalCount:   group.TotalCount,
    }
  }
  
  enc := json.NewEncoder(w)
  result := KeywordGroupResults{Groups: responseGroups}
  enc.Encode(&result)
}

// UserQueriesResult represents the response for user queries endpoint
type UserQueriesResult struct {
  User    string                `json:"user"`
  Queries []QueryWithTimestamp `json:"queries"`
}

// AllUserQueriesResult represents the response for all user queries endpoint
type AllUserQueriesResult struct {
  Users map[string][]QueryWithTimestamp `json:"users"`
}

// HandleUserQueries handles requests for user query history from logs
func HandleUserQueries(w http.ResponseWriter, r *http.Request, db *Database, logDir string) {
  if logDir == "" {
    http.Error(w, "Log directory not configured", http.StatusInternalServerError)
    return
  }
  
  // Get user email from context
  userEmail := r.Context().Value("userEmail").(string)
  log.Printf("User queries request from %s", userEmail)
  
  // Restrict access to specific admin user only
  if userEmail != "matthieu.devin@gmail.com" {
    log.Printf("Access denied for user queries endpoint: %s", userEmail)
    http.Error(w, "Access denied", http.StatusForbidden)
    return
  }
  
  // Create log parser
  parser := NewLogParser(logDir)
  
  // Check if requesting specific user queries or all users
  requestedUser := r.FormValue("user")
  
  if requestedUser != "" {
    // Get queries for specific user
    queries, err := parser.GetQueriesForUser(requestedUser)
    if err != nil {
      log.Printf("Error getting queries for user %s: %v", requestedUser, err)
      http.Error(w, "Error retrieving user queries", http.StatusInternalServerError)
      return
    }
    
    result := UserQueriesResult{
      User:    requestedUser,
      Queries: queries,
    }
    
    enc := json.NewEncoder(w)
    enc.Encode(&result)
  } else {
    // Get all user queries
    allQueries, err := parser.GroupQueriesByUser()
    if err != nil {
      log.Printf("Error getting all user queries: %v", err)
      http.Error(w, "Error retrieving user queries", http.StatusInternalServerError)
      return
    }
    
    result := AllUserQueriesResult{
      Users: allQueries,
    }
    
    enc := json.NewEncoder(w)
    enc.Encode(&result)
  }
}

