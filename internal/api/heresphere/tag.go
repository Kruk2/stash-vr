package heresphere

import (
	"fmt"
	"sort"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/config"
	"stash-vr/internal/stash/gql"
	"stash-vr/internal/util"
)

type tag struct {
	Name   string  `json:"name"`
	Start  float64 `json:"start"`
	End    float64 `json:"end"`
	Track  *int    `json:"track,omitempty"`
	Rating float32 `json:"rating"`
}

const seperator = ":"

func getTags(s gql.SceneScanParts) []tag {
	var tagTracks [][]tag

	markers := getMarkers(s)
	performers := getPerformers(s)
	studio := getStudio(s)
	stashTags := getStashTags(s)

	meta := make([]tag, 0, len(studio)+len(performers))
	meta = append(meta, studio...)
	meta = append(meta, performers...)

	fillTagDurations(markers)
	duration := s.Files[0].Duration * 1000
	equallyDivideTagDurations(duration, meta)

	tagTracks = append(tagTracks, markers)
	for i := range stashTags {
		stashTags[i].Start = 0
		stashTags[i].End = duration
		tmp := make([]tag, 0, 1)
		tmp = append(tmp, stashTags[i])
		tagTracks = append(tagTracks, tmp)
	}

	tagTracks = append(tagTracks, meta)

	track := 0
	tags := make([]tag, 0, len(tagTracks))
	for i := range tagTracks {
		if len(tagTracks[i]) == 0 {
			continue
		}
		for j := range tagTracks[i] {
			tagTracks[i][j].Track = util.Ptr(track)
			tags = append(tags, tagTracks[i][j])
		}
		track++
	}
	return tags
}

func getPerformers(s gql.SceneScanParts) []tag {
	tags := make([]tag, len(s.Performers))
	for i, p := range s.Performers {
		tags[i] = tag{
			Name:   internal.LegendPerformer.Full + seperator + p.Name,
			Rating: float32(p.Rating100) / 20.0,
		}
	}
	return tags
}

func getStudio(s gql.SceneScanParts) []tag {
	if s.Studio == nil {
		return nil
	}
	return []tag{{
		Name:   internal.LegendStudio.Full + seperator + s.Studio.Name,
		Rating: float32(s.Studio.Rating100) / 20.0,
	}}
}

func getFields(s gql.SceneScanParts) []tag {
	tags := []tag{
		{Name: fmt.Sprintf("%s:%d", internal.LegendPlayCount.Short, s.Play_count)},
		{Name: fmt.Sprintf("%s:%d", internal.LegendOCount.Short, s.O_counter)},
		{Name: fmt.Sprintf("%s:%v", internal.LegendOrganized.Short, s.Organized)}}

	return tags
}

func getStashTags(s gql.SceneScanParts) []tag {
	tags := make([]tag, 0, len(s.Tags))
	for _, t := range s.Tags {
		if t.Name == config.Get().FavoriteTag {
			continue
		}
		t := tag{
			Name: internal.LegendTag.Short + seperator + t.Name,
		}
		tags = append(tags, t)
	}
	return tags
}

func getMarkers(s gql.SceneScanParts) []tag {
	tags := make([]tag, len(s.Scene_markers))
	for i, sm := range s.Scene_markers {
		tagName := sm.Primary_tag.Name
		t := tag{
			Name:  internal.LegendTag.Short + seperator + tagName,
			Start: sm.Seconds * 1000,
		}
		tags[i] = t
	}
	return tags
}

func equallyDivideTagDurations(totalDuration float64, tags []tag) {
	durationPerItem := totalDuration / float64(len(tags))
	for i := range tags {
		tags[i].Start = float64(i) * durationPerItem
		tags[i].End = float64(i+1) * durationPerItem
	}
}

func fillTagDurations(tags []tag) {
	sort.Slice(tags, func(i, j int) bool { return tags[i].Start < tags[j].Start })
	for i := range tags {
		if i == len(tags)-1 {
			tags[i].End = 0
		} else if tags[i+1].Start == 0 {
			tags[i].End = tags[i].Start + 20000
		} else {
			tags[i].End = tags[i+1].Start
		}
	}
}
