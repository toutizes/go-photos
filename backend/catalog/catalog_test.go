package catalog

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"toutizes.com/go-photwo/backend/store"
)

func TestDropAccentsAndSynonyms(t *testing.T) {
	cache := make(map[string]string)
	if got := DropAccents("Hélène & François à l'École ÇA", cache); got != "helene & francois a l'ecole ca" {
		t.Fatalf("unexpected DropAccents result: %q", got)
	}
	if got := DropAccents("Hélène & François à l'École ÇA", cache); got != "helene & francois a l'ecole ca" {
		t.Fatalf("unexpected cached DropAccents result: %q", got)
	}

	dir := t.TempDir()
	synsContent := "julien devin,juju\ncoline devin,coco\n"
	if err := os.WriteFile(filepath.Join(dir, "synonyms.txt"), []byte(synsContent), 0644); err != nil {
		t.Fatal(err)
	}
	syns := LoadSynonyms(dir)
	if !syns.IsName("juju") || !syns.IsName("julien devin") || syns.IsName("unknown") {
		t.Fatalf("unexpected IsName results")
	}
	p := syns.Lookup("juju")
	if p == nil || len(p.Names) != 2 {
		t.Fatalf("expected person with 2 aliases, got %+v", p)
	}
}

func TestMergeImageAndUpdateDirectory(t *testing.T) {
	its123 := int64(123)
	oldImg := &store.Item{Keywords: []string{"a", "b"}, ItemTimestamp: &its123}

	mrg := mergeImage(oldImg, &store.Item{})
	if len(mrg.Keywords) != 2 || *mrg.ItemTimestamp != 123 {
		t.Fatalf("unexpected mergeImage result: %+v", mrg)
	}

	its456 := int64(456)
	mrg = mergeImage(oldImg, &store.Item{ItemTimestamp: &its456})
	if len(mrg.Keywords) != 2 || *mrg.ItemTimestamp != 456 {
		t.Fatalf("unexpected mergeImage result with new timestamp: %+v", mrg)
	}

	its789 := int64(789)
	mrg = mergeImage(oldImg, &store.Item{Keywords: []string{"z"}, ItemTimestamp: &its789})
	if len(mrg.Keywords) != 1 || mrg.Keywords[0] != "z" || *mrg.ItemTimestamp != 789 {
		t.Fatalf("unexpected mergeImage result with new keywords: %+v", mrg)
	}

	dir := t.TempDir()
	for _, f := range []string{"foo.jpg", "bar.webm", "fee", "gee.JPG", "bidon"} {
		if err := os.WriteFile(filepath.Join(dir, f), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	subs, err := readDirFileInfos(dir)
	if err != nil {
		t.Fatal(err)
	}

	sdir := store.Directory{
		Items: []*store.Item{
			{Name: proto.String("old.img"), Keywords: []string{"baba"}, Image: &store.Image{}},
			{Name: proto.String("gee.JPG"), Keywords: []string{"ba ba"}, Image: &store.Image{}},
		},
	}
	if err := UpdateDirectory(dir, subs, false, &sdir); err != nil {
		t.Fatal(err)
	}
	if len(sdir.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(sdir.Items))
	}
	foundGee := false
	for _, it := range sdir.Items {
		if it.GetName() == "gee.JPG" {
			foundGee = true
			if len(it.Keywords) != 1 || it.Keywords[0] != "ba ba" {
				t.Fatalf("expected gee.JPG to preserve keyword 'ba ba', got %v", it.Keywords)
			}
		}
	}
	if !foundGee {
		t.Fatal("missing gee.JPG")
	}
}

func TestQueriesAndActivity(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "synonyms.txt"), []byte("julien devin,juju\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cat := NewCatalog("/tmp/orig", "/tmp/root", staticDir)
	cat.Synonyms = LoadSynonyms(staticDir)

	t1 := time.Date(2024, 7, 14, 12, 0, 0, 0, pstZone)
	t2 := time.Date(2025, 6, 15, 15, 0, 0, 0, pstZone)
	t3 := time.Date(2025, 6, 18, 15, 0, 0, 0, pstZone)

	a1 := &Album{
		RelPat:       "2024/2024-07-14 Paris",
		IndexTime:    t1,
		LastModified: t1,
	}
	p1 := &Photo{
		Album:       a1,
		Name:        "bastille.jpg",
		Keywords:    []string{"julien devin", "paris", "été"},
		SubKeywords: []string{"julien", "devin"},
		FileTime:    t1,
		ItemTime:    t1,
		Width:       1600,
		Height:      1200,
		Stereo:      &Stereo{Dx: 5, Dy: -2},
	}
	a1.Photos = []*Photo{p1}

	a2 := &Album{
		RelPat:       "2025/2025-06-15 Beach",
		IndexTime:    t3,
		LastModified: t3,
	}
	p2 := &Photo{
		Album:       a2,
		Name:        "beach1.jpg",
		Keywords:    []string{"vacances", "plage", "famille"},
		SubKeywords: nil,
		FileTime:    t2,
		ItemTime:    t2,
		Width:       2000,
		Height:      1500,
	}
	p3 := &Photo{
		Album:       a2,
		Name:        "beach2.jpg",
		Keywords:    []string{"vacances", "plage", "juju"},
		SubKeywords: nil,
		FileTime:    t3,
		ItemTime:    t3,
		Width:       2000,
		Height:      1500,
	}
	a2.Photos = []*Photo{p2, p3}

	cat.Albums = []*Album{a1, a2}
	cat.Index.internAlbum(a1)
	cat.Index.internAlbum(a2)
	if n := cat.Index.Build(cat.Albums); n != 3 {
		t.Fatalf("expected 3 indexed photos, got %d", n)
	}

	// Test synonym expansion (multi-word and single-word nickname)
	if res := cat.QueryPhotos("julien devin"); len(res) != 2 {
		t.Fatalf("expected 2 photos for 'julien devin' via synonyms, got %d", len(res))
	}
	if res := cat.QueryPhotos("juju"); len(res) != 2 {
		t.Fatalf("expected 2 photos for 'juju' via synonyms, got %d", len(res))
	}

	// Test accent-insensitive search
	if res := cat.QueryPhotos("ete"); len(res) != 1 || res[0] != p1 {
		t.Fatalf("expected p1 for accent-stripped 'ete', got %v", res)
	}

	// Test comma-separated AND query
	if res := cat.QueryPhotos("vacances, 2025-06-18"); len(res) != 1 || res[0] != p3 {
		t.Fatalf("expected p3 for 'vacances, 2025-06-18', got %v", res)
	}

	// Test year range, month-day, album, titre, stereo, count, albums:
	if res := cat.QueryPhotos("2024--2025"); len(res) != 3 {
		t.Fatalf("expected 3 photos for 2024--2025, got %d", len(res))
	}
	if res := cat.QueryPhotos("07-14"); len(res) != 1 || res[0] != p1 {
		t.Fatalf("expected p1 for 07-14, got %v", res)
	}
	if res := cat.QueryPhotos("album:2025/2025-06-15 Beach"); len(res) != 2 {
		t.Fatalf("expected 2 photos for exact album:, got %d", len(res))
	}
	if res := cat.QueryPhotos("titre:paris"); len(res) != 1 || res[0] != p1 {
		t.Fatalf("expected p1 for titre:paris, got %v", res)
	}
	if res := cat.QueryPhotos("stereo:"); len(res) != 1 || res[0] != p1 {
		t.Fatalf("expected p1 for stereo:, got %v", res)
	}
	if res := cat.QueryPhotos("count:3"); len(res) != 3 {
		t.Fatalf("expected 3 photos with 3 keywords, got %d", len(res))
	}
	if res := cat.QueryPhotos("albums:"); len(res) != 2 {
		t.Fatalf("expected 2 album representative photos for albums:, got %d", len(res))
	}

	// Test recent keyword groups
	groups := ComputeRecentKeywordGroups(cat.Albums)
	if len(groups) != 1 {
		t.Fatalf("expected 1 keyword group (vacances + plage), got %d", len(groups))
	}
	if len(groups[0].Keywords) != 2 || groups[0].TotalCount != 2 {
		t.Fatalf("unexpected group contents: %+v", groups[0])
	}
}

func TestLoadAndMinifyCatalog(t *testing.T) {
	origRoot := t.TempDir()
	root := t.TempDir()
	albumDir := filepath.Join(origRoot, "2025-05-01")
	if err := os.MkdirAll(albumDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a synthetic 800x600 JPEG
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	imgPath := filepath.Join(albumDir, "sample.jpg")
	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	cat, err := LoadCatalog(Options{
		OrigRoot:        origRoot,
		Root:            root,
		UpdateDisk:      true,
		Minify:          true,
		MinifierThreads: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Albums) != 2 { // "" and "2025-05-01"
		t.Fatalf("expected 2 directories, got %d", len(cat.Albums))
	}
	res := cat.QueryPhotos("sample")
	if len(res) != 1 {
		t.Fatalf("expected 1 photo matching filename 'sample', got %d", len(res))
	}
	p := res[0]
	if p.Width != 800 || p.Height != 600 {
		t.Fatalf("expected 800x600 dimensions, got %dx%d", p.Width, p.Height)
	}

	// Verify index.pbin, mini, and midi were written
	if _, err := os.Stat(cat.IndexPath("2025-05-01")); err != nil {
		t.Fatalf("missing index.pbin: %v", err)
	}
	miniCfg, err := os.Open(cat.MiniPath(p))
	if err != nil {
		t.Fatalf("missing mini file: %v", err)
	}
	defer miniCfg.Close()
	cfg, _, err := image.DecodeConfig(miniCfg)
	if err != nil || cfg.Width != 360 || cfg.Height != 360 {
		t.Fatalf("expected 360x360 mini thumbnail, got %+v (err=%v)", cfg, err)
	}
}
