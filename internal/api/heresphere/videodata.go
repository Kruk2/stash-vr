package heresphere

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"stash-vr/internal/api/heatmap"
	"stash-vr/internal/config"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
)

type videoData struct {
	Access int `json:"access"`

	Title          string   `json:"title"`
	Description    string   `json:"description"`
	ThumbnailImage string   `json:"thumbnailImage"`
	ThumbnailVideo string   `json:"thumbnailVideo"`
	DateReleased   string   `json:"dateReleased"`
	DateAdded      string   `json:"dateAdded"`
	Duration       float64  `json:"duration"`
	Rating         float32  `json:"rating"`
	Favorites      int      `json:"favorites"`
	IsFavorite     bool     `json:"isFavorite"`
	Projection     string   `json:"projection"`
	Stereo         string   `json:"stereo"`
	Fov            float32  `json:"fov"`
	Lens           string   `json:"lens"`
	Scripts        []script `json:"scripts"`
	Tags           []tag    `json:"tags"`
	Media          []media  `json:"media"`

	WriteFavorite bool `json:"writeFavorite"`
	WriteRating   bool `json:"writeRating"`
	WriteTags     bool `json:"writeTags"`
}

type media struct {
	Name    string   `json:"name"`
	Sources []source `json:"sources"`
}

type source struct {
	Resolution int    `json:"resolution"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	Size       int    `json:"size"`
	Url        string `json:"url"`
}

type script struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func buildVideoData(ctx context.Context, client graphql.Client, baseUrl string, sceneId string) (videoData, error) {
	findSceneResponse, err := gql.FindSceneFull(ctx, client, sceneId)
	if err != nil {
		return videoData{}, fmt.Errorf("FindSceneFull: %w", err)
	}
	if findSceneResponse.FindScene == nil {
		return videoData{}, fmt.Errorf("FindSceneFull: not found")
	}
	s := findSceneResponse.FindScene.SceneFullParts

	if len(s.SceneScanParts.Files) == 0 {
		return videoData{}, fmt.Errorf("scene %s has no files", sceneId)
	}

	thumbnailUrl := stash.ApiKeyed(s.Paths.Screenshot)
	if !config.Get().IsHeatmapDisabled && s.ScriptParts.Interactive && s.ScriptParts.Paths.Interactive_heatmap != "" {
		thumbnailUrl = heatmap.GetCoverUrl(baseUrl, sceneId)
	}

	title := s.Title
	if title == "" {
		title = s.SceneScanParts.Files[0].Basename
	}

	vd := videoData{
		Access:         1,
		Title:          title,
		Description:    s.Details,
		ThumbnailImage: thumbnailUrl,
		ThumbnailVideo: stash.ApiKeyed(s.Paths.Preview),
		DateReleased:   s.Date,
		DateAdded:      s.Created_at.Format("2006-01-02"),
		Duration:       s.SceneScanParts.Files[0].Duration * 1000,
		Rating:         float32(s.Rating100) / 20.0,
		Favorites:      s.O_counter,
		WriteFavorite:  true,
		WriteRating:    true,
		WriteTags:      true,
	}

	setIsFavorite(s, &vd)
	setStreamSources(ctx, s, &vd)
	set3DFormat(s, &vd)
	setTags(s, &vd)

	setScripts(s, &vd)
	return vd, nil
}

func setTags(s gql.SceneFullParts, videoData *videoData) {
	tags := getTags(s.SceneScanParts)
	videoData.Tags = tags
}

func setScripts(s gql.SceneFullParts, videoData *videoData) {
	if s.ScriptParts.Interactive {
		videoData.Scripts = append(videoData.Scripts, script{
			Name: "Script-" + s.Title,
			Url:  stash.ApiKeyed(s.ScriptParts.Paths.Funscript),
		})
	}
}

func ContainsI(a string, b string) bool {
	return strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)
}

func set3DFormat(s gql.SceneFullParts, videoData *videoData) {
	videoData.Projection = "equirectangular"
	videoData.Stereo = "sbs"
	videoData.Fov = 180.0
	videoData.Lens = "Linear"

	isVr := strings.Contains(videoData.Media[0].Sources[0].Url, "/VR/") || strings.Contains(videoData.Media[0].Sources[0].Url, "\\VR\\")
	filenameSeparator := regexp.MustCompile("[ _.-]+")

	nameparts := filenameSeparator.Split(strings.ToLower(filepath.Base(videoData.Media[0].Sources[0].Url)), -1)
	for i, part := range nameparts {
		videoProjection := ""

		if !isVr {
			videoProjection = "flat"
		} else if part == "mkx200" || part == "mkx220" || part == "rf52" || part == "fisheye190" || part == "vrca220" || part == "flat" {
			videoProjection = part
		} else if part == "fisheye" || part == "f180" || part == "180f" {
			videoProjection = "fisheye"
		} else if i < len(nameparts)-1 && (part+"_"+nameparts[i+1] == "mono_360" || part+"_"+nameparts[i+1] == "mono_180") {
			videoProjection = nameparts[i+1] + "_mono"
		} else if i < len(nameparts)-1 && (part+"_"+nameparts[i+1] == "360_mono" || part+"_"+nameparts[i+1] == "180_mono") {
			videoProjection = part + "_mono"
		} else {
			continue
		}

		switch videoProjection {
		case "flat":
			videoData.Projection = "perspective"
			videoData.Stereo = "mono"
			videoData.Fov = 0

		case "180_mono":
			videoData.Projection = "equirectangular"
			videoData.Stereo = "mono"

		case "360_mono":
			videoData.Projection = "equirectangular360"
			videoData.Stereo = "mono"

		case "180_sbs":
			videoData.Projection = "equirectangular"

		case "360_tb":
			videoData.Projection = "equirectangular360"
			videoData.Stereo = "tb"

		case "mkx200":
			videoData.Projection = "fisheye"
			videoData.Fov = 200.0
			videoData.Lens = "MKX200"

		case "mkx220":
			videoData.Projection = "fisheye"
			videoData.Fov = 220.0
			videoData.Lens = "MKX220"

		case "vrca220":
			videoData.Projection = "fisheye"
			videoData.Fov = 220.0
			videoData.Lens = "VRCA220"

		case "rf52":
			videoData.Projection = "fisheye"
			videoData.Fov = 190.0

		case "fisheye190":
			videoData.Projection = "fisheye"
			videoData.Fov = 190.0

		case "fisheye":
			videoData.Projection = "fisheye"
		}
	}
}

func setStreamSources(ctx context.Context, s gql.SceneFullParts, videoData *videoData) {
	for _, stream := range stash.GetStreams(ctx, s.StreamsParts, true) {
		e := media{
			Name: stream.Name,
		}
		if len(stream.Sources) != 1 {
			log.Ctx(ctx).Error().Msg("Wrong number of sources for scene " + s.Id)
		}
		for _, s := range stream.Sources {
			if ContainsI(s.Url, "TOPAZ") {
				vs := source{
					Resolution: s.Resolution,
					Url:        s.Url,
				}
				e.Name = "Topaz"
				e.Sources = append(e.Sources, vs)

			} else {
				vs := source{
					Resolution: s.Resolution,
					Url:        s.Url,
				}
				if strings.Index(s.Url, "http") == 0 {
					vs.Url = "file://" + vs.Url
				}
				e.Name = filepath.Base(s.Url)
				e.Sources = append(e.Sources, vs)
			}
		}
		videoData.Media = append(videoData.Media, e)
	}
}

func setIsFavorite(s gql.SceneFullParts, videoData *videoData) {
	videoData.IsFavorite = ContainsFavoriteTag(s.TagPartsArray)
}

func ContainsFavoriteTag(ts gql.TagPartsArray) bool {
	for _, t := range ts.Tags {
		if t.Name == config.Get().FavoriteTag {
			return true
		}
	}
	return false
}
