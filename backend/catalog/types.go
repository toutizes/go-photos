package catalog

import (
	"time"

	"google.golang.org/protobuf/proto"
	"toutizes.com/go-photwo/backend/store"
)

// Stereo holds stereoscopic alignment offsets for a photo.
type Stereo struct {
	Dx    float32
	Dy    float32
	AnaDx float32 `json:"-"`
	AnaDy float32 `json:"-"`
}

// Photo represents an indexed photo in an Album.
type Photo struct {
	Album         *Album
	Name          string
	Keywords      []string
	SubKeywords   []string
	FileTime      time.Time
	ItemTime      time.Time
	Height        int32
	Width         int32
	RotateDegrees int32
	Stereo        *Stereo
	Id            int
	Rank          int
}

// Album represents a photo directory relative to orig_root.
type Album struct {
	RelPat       string
	IndexTime    time.Time
	LastModified time.Time
	OrderBy      string
	Photos       []*Photo
}

// ItemTime returns the capture timestamp of the first photo in the album,
// falling back to IndexTime if the album has no photos.
func (a *Album) ItemTime() time.Time {
	if len(a.Photos) == 0 {
		return a.IndexTime
	}
	return a.Photos[0].ItemTime
}

// ToProto converts a Photo back to its store.Item protobuf representation.
func (p *Photo) ToProto() *store.Item {
	item := &store.Item{
		Name:     proto.String(p.Name),
		Keywords: p.Keywords,
		Image: &store.Image{
			Height: proto.Int32(p.Height),
			Width:  proto.Int32(p.Width),
		},
	}
	if !p.FileTime.Equal(time.Unix(0, 0)) {
		item.FileTimestamp = proto.Int64(TimeToProto(p.FileTime))
	}
	if !p.ItemTime.Equal(time.Unix(0, 0)) {
		item.ItemTimestamp = proto.Int64(TimeToProto(p.ItemTime))
	}
	if p.RotateDegrees != 0 {
		item.Image.RotateDegrees = proto.Int32(p.RotateDegrees)
	}
	if p.Stereo != nil {
		item.Image.Stereo = &store.Stereo{
			Dx:    proto.Float32(p.Stereo.Dx),
			Dy:    proto.Float32(p.Stereo.Dy),
			AnaDx: proto.Float32(p.Stereo.AnaDx),
			AnaDy: proto.Float32(p.Stereo.AnaDy),
		}
	}
	return item
}

// ToProto converts an Album to its store.Directory protobuf representation.
func (a *Album) ToProto() *store.Directory {
	sdir := &store.Directory{}
	if !a.IndexTime.Equal(time.Unix(0, 0)) {
		sdir.DirectoryTimestamp = proto.Int64(TimeToProto(a.IndexTime))
	}
	sdir.Items = make([]*store.Item, len(a.Photos))
	for i, p := range a.Photos {
		sdir.Items[i] = p.ToProto()
	}
	return sdir
}

// JsonImage is the wire format returned by /db/q for photo searches.
type JsonImage struct {
	Id     int
	Ad     string
	In     string
	Its    int64
	Fts    int64
	H      int32
	W      int32
	Kwd    []string
	Stereo *Stereo
}

// FillJson populates a JsonImage from a Photo.
func (p *Photo) FillJson(dst *JsonImage) {
	dst.Id = p.Id
	dst.Ad = p.Album.RelPat
	dst.In = p.Name
	dst.Its = p.ItemTime.Unix()
	dst.Fts = p.FileTime.Unix()
	dst.H = p.Height
	dst.W = p.Width
	dst.Kwd = p.Keywords
	dst.Stereo = p.Stereo
}

// JsonDirectory is the wire format returned by /db/q?kind=album.
type JsonDirectory struct {
	Id    string
	Ats   int64
	Dts   int64
	Nimgs int
	Cov   int
}

// FillJson populates a JsonDirectory from an Album.
func (a *Album) FillJson(dst *JsonDirectory) {
	dst.Id = a.RelPat
	if len(a.Photos) > 0 {
		dst.Ats = a.Photos[0].ItemTime.Unix()
		dst.Cov = a.Photos[0].Id
	} else {
		dst.Ats = a.IndexTime.Unix()
	}
	dst.Dts = a.LastModified.Unix()
	dst.Nimgs = len(a.Photos)
}

// ImageInfo represents simplified photo metadata in recent keyword responses.
type ImageInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// KeywordResponse represents a keyword and its recent sample images.
type KeywordResponse struct {
	Keyword      string      `json:"keyword"`
	Count        int         `json:"count"`
	RecentImages []ImageInfo `json:"-"`
}

// KeywordGroupResponse represents a group of keywords sharing the same top image.
type KeywordGroupResponse struct {
	Keywords     []KeywordResponse `json:"keywords"`
	RecentImages []ImageInfo       `json:"recent_images"`
	TotalWeight  float64           `json:"-"`
	TotalCount   int               `json:"total_count"`
}

// KeywordGroupResults is the top-level JSON payload for /db/recent-keyword-groups.
type KeywordGroupResults struct {
	Groups []KeywordGroupResponse `json:"groups"`
}

// ProtoToTime converts a millisecond proto timestamp to time.Time.
func ProtoToTime(ptime int64) time.Time {
	return time.Unix(ptime/1000, 0)
}

// TimeToProto converts a time.Time to a millisecond proto timestamp.
func TimeToProto(t time.Time) int64 {
	return 1000 * t.Unix()
}
