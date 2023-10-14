package deovr

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
)

type videoData struct {
	Authorized     string              `json:"authorized"`
	FullAccess     bool                `json:"fullAccess"`
	Title          string              `json:"title"`
	Id             string              `json:"id"`
	VideoLength    int                 `json:"videoLength"`
	Is3d           bool                `json:"is3d"`
	ScreenType     string              `json:"screenType"`
	StereoMode     string              `json:"stereoMode"`
	SkipIntro      int                 `json:"skipIntro"`
	VideoThumbnail string              `json:"videoThumbnail,omitempty"`
	VideoPreview   string              `json:"videoPreview,omitempty"`
	ThumbnailUrl   string              `json:"thumbnailUrl"`
	ChromaKey      *videoDataChromaKey `json:"chromaKey"`

	TimeStamps []timeStamp `json:"timeStamps,omitempty"`

	Encodings []encoding `json:"encodings"`
}

type videoDataChromaKey struct {
	HasAlpha bool `json:"hasAlpha"`
}

type timeStamp struct {
	Ts   int    `json:"ts"`
	Name string `json:"name"`
}

type encoding struct {
	Name         string        `json:"name"`
	VideoSources []videoSource `json:"videoSources"`
}

type videoSource struct {
	Resolution int    `json:"resolution"`
	Url        string `json:"url"`
}

func buildVideoData(ctx context.Context, client graphql.Client, baseUrl string, sceneId string) (videoData, error) {
	findSceneResponse, err := gql.FindSceneFull(ctx, client, sceneId)
	if err != nil {
		return videoData{}, fmt.Errorf("FindScene: %w", err)
	}
	if findSceneResponse.FindScene == nil {
		return videoData{}, fmt.Errorf("FindScene: not found")
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
		Authorized:   "1",
		FullAccess:   true,
		Title:        title,
		Id:           s.Id,
		VideoLength:  int(s.SceneScanParts.Files[0].Duration),
		SkipIntro:    0,
		VideoPreview: stash.ApiKeyed(s.Paths.Preview),
		ThumbnailUrl: thumbnailUrl,
	}

	setChromaKey(findSceneResponse, &vd)
	setStreamSources(ctx, s, &vd)
	setMarkers(s, &vd)
	set3DFormat(s, &vd)

	return vd, nil
}

func setChromaKey(findSceneResponse *gql.FindSceneFullResponse, videoData *videoData) {
	tagName := config.Get().PassThroughTag

	if ContainsTag(findSceneResponse.FindScene.TagPartsArray, tagName) {
		videoData.ChromaKey = &videoDataChromaKey{HasAlpha: true}
	} else {
		videoData.ChromaKey = nil
	}
}

func setStreamSources(ctx context.Context, s gql.SceneFullParts, videoData *videoData) {
	streams := stash.GetStreams(ctx, s.StreamsParts, false)
	videoData.Encodings = make([]encoding, len(streams))
	for i, stream := range streams {
		videoData.Encodings[i] = encoding{
			Name:         stream.Name,
			VideoSources: make([]videoSource, len(stream.Sources)),
		}
		for j, source := range stream.Sources {
			videoData.Encodings[i].VideoSources[j] = videoSource{
				Resolution: source.Resolution,
				Url:        source.Url,
			}
		}
	}
}

func setMarkers(s gql.SceneFullParts, videoData *videoData) {
	for _, sm := range s.Scene_markers {
		sb := strings.Builder{}
		sb.WriteString(sm.Primary_tag.Name)
		if sm.Title != "" {
			sb.WriteString(":")
			sb.WriteString(sm.Title)
		}
		ts := timeStamp{
			Ts:   int(sm.Seconds),
			Name: sb.String(),
		}
		videoData.TimeStamps = append(videoData.TimeStamps, ts)
	}
}

func ContainsI(a string, b string) bool {
	return strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)
}

func set3DFormat(s gql.SceneFullParts, videoData *videoData) {
	isVr := strings.Contains(videoData.Encodings[0].VideoSources[0].Url, "/VR/") || strings.Contains(videoData.Encodings[0].VideoSources[0].Url, "\\VR\\")
	if !isVr {
		return
	}

	videoData.ScreenType = "dome"
	videoData.Is3d = true
	videoData.StereoMode = "sbs"

	filenameSeparator := regexp.MustCompile("[ _.-]+")

	nameparts := filenameSeparator.Split(strings.ToLower(filepath.Base(videoData.Encodings[0].VideoSources[0].Url)), -1)
	for i, part := range nameparts {
		videoProjection := ""

		if part == "mkx200" || part == "mkx220" || part == "rf52" || part == "fisheye190" || part == "vrca220" || part == "flat" {
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
			videoData.ScreenType = "flat"

		case "180_mono":
			videoData.ScreenType = "dome"
			videoData.StereoMode = "mono"

		case "360_mono":
			videoData.ScreenType = "sphere"
		case "180_sbs":
			videoData.ScreenType = "dome"
		case "360_tb":
			videoData.ScreenType = "sphere"
			videoData.StereoMode = "tb"

		case "mkx200":
			videoData.ScreenType = "mkx200"

		case "mkx220":
			videoData.ScreenType = "mkx220"

		case "vrca220":
			videoData.ScreenType = "vrca220"

		case "rf52":
			videoData.ScreenType = "rf52"

		case "fisheye190":
			videoData.ScreenType = "fisheye190"

		case "fisheye":
			videoData.ScreenType = "fisheye"
		}
	}

	videoData.Title = videoData.ScreenType + " - " + videoData.Title
	videoData.ScreenType = ""
	videoData.StereoMode = ""
	videoData.Is3d = false
}

func ContainsTag(tagPartsArray gql.TagPartsArray, s string) bool {
	for _, t := range tagPartsArray.Tags {
		if t.Name == s {
			return true
		}
	}
	return false
}
