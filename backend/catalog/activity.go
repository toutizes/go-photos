package catalog

import (
	"sort"
)

type keywordCount struct {
	Keyword      string
	Weight       float64
	Count        int
	RecentImages []*Photo
}

// ComputeRecentKeywordGroups computes the weighted keyword groups across the 20 most recent
// albums (ordered by Album.ItemTime descending), grouping keywords that share the same
// most recent photo.
func ComputeRecentKeywordGroups(albums []*Album) []KeywordGroupResponse {
	if len(albums) == 0 {
		return []KeywordGroupResponse{}
	}

	sortedAlbums := make([]*Album, len(albums))
	copy(sortedAlbums, albums)
	sort.Slice(sortedAlbums, func(i, j int) bool {
		return sortedAlbums[i].ItemTime().After(sortedAlbums[j].ItemTime())
	})
	if len(sortedAlbums) > 20 {
		sortedAlbums = sortedAlbums[:20]
	}

	currentTime := sortedAlbums[0].ItemTime()
	keywordImages := make(map[string][]*Photo)
	keywordWeights := make(map[string]float64)

	for _, album := range sortedAlbums {
		for _, img := range album.Photos {
			for _, keyword := range img.Keywords {
				if keyword == "" {
					continue
				}
				keywordImages[keyword] = append(keywordImages[keyword], img)
				if len(keywordImages[keyword]) <= 5 {
					since := currentTime.Sub(img.ItemTime).Hours()
					if since < 1.0 {
						since = 1.0
					}
					keywordWeights[keyword] += 1.0 / since
				}
			}
		}
	}

	keywordCounts := make([]keywordCount, 0, len(keywordImages))
	for keyword, images := range keywordImages {
		if len(images) < 2 {
			continue
		}
		sort.Slice(images, func(i, j int) bool {
			return images[i].ItemTime.After(images[j].ItemTime)
		})
		maxImages := 4
		if len(images) < maxImages {
			maxImages = len(images)
		}
		keywordCounts = append(keywordCounts, keywordCount{
			Keyword:      keyword,
			Weight:       keywordWeights[keyword],
			Count:        len(images),
			RecentImages: images[:maxImages],
		})
	}

	sort.Slice(keywordCounts, func(i, j int) bool {
		if keywordCounts[i].Weight == keywordCounts[j].Weight {
			return keywordCounts[i].Keyword < keywordCounts[j].Keyword
		}
		return keywordCounts[i].Weight > keywordCounts[j].Weight
	})

	imageIDToKeywords := make(map[int][]keywordCount)
	imageIDToImages := make(map[int][]*Photo)
	for _, kc := range keywordCounts {
		firstID := kc.RecentImages[0].Id
		imageIDToKeywords[firstID] = append(imageIDToKeywords[firstID], kc)
		if imageIDToImages[firstID] == nil {
			imageIDToImages[firstID] = kc.RecentImages
		}
	}

	groups := make([]KeywordGroupResponse, 0, len(imageIDToKeywords))
	for imageID, kwds := range imageIDToKeywords {
		sort.Slice(kwds, func(i, j int) bool {
			if kwds[i].Weight == kwds[j].Weight {
				return kwds[i].Keyword < kwds[j].Keyword
			}
			return kwds[i].Weight > kwds[j].Weight
		})
		if len(kwds) > 5 {
			kwds = kwds[:5]
		}

		totalWeight := 0.0
		allImageIDs := make(map[int]bool)
		respKwds := make([]KeywordResponse, len(kwds))
		for i, kc := range kwds {
			totalWeight += kc.Weight
			imgs := make([]ImageInfo, len(kc.RecentImages))
			for j, img := range kc.RecentImages {
				allImageIDs[img.Id] = true
				imgs[j] = ImageInfo{Id: img.Id, Name: img.Name}
			}
			respKwds[i] = KeywordResponse{
				Keyword:      kc.Keyword,
				Count:        kc.Count,
				RecentImages: imgs,
			}
		}

		groupPhotos := imageIDToImages[imageID]
		groupImgs := make([]ImageInfo, len(groupPhotos))
		for i, img := range groupPhotos {
			groupImgs[i] = ImageInfo{Id: img.Id, Name: img.Name}
		}

		groups = append(groups, KeywordGroupResponse{
			Keywords:     respKwds,
			RecentImages: groupImgs,
			TotalWeight:  totalWeight,
			TotalCount:   len(allImageIDs),
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].TotalWeight > groups[j].TotalWeight
	})
	return groups
}
