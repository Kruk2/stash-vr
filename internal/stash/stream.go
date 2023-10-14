package stash

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"stash-vr/internal/stash/gql"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type Stream struct {
	Name    string
	Sources []Source
}

type Source struct {
	Resolution int
	Url        string
}

var rgxResolution = regexp.MustCompile(`\((\d+)p\)`)

func GetStreams(ctx context.Context, fsp gql.StreamsParts, sortResolutionAsc bool) []Stream {
	streams := make([]Stream, len(fsp.Files))
	isVR := false

	for i, file := range fsp.Files {
		streams[i] = Stream{
			Name: strings.TrimSuffix(file.Basename, filepath.Ext(file.Basename)),
			Sources: []Source{{
				Resolution: file.Height,
				Url:        strings.Replace(file.Path, "\\", "/", -1),
			}},
		}

		if strings.Contains(file.Path, "\\VR\\") || strings.Contains(file.Path, "/VR/") {
			isVR = true
		}
	}

	sort.Slice(streams, func(i, j int) bool {
		return streams[i].Sources[0].Url < streams[j].Sources[0].Url
	})

	if isVR {
		for _, file := range fsp.SceneStreams {
			if file.Label == "MP4 Topaz" {
				streams = append(streams, Stream{
					Name: file.Label,
					Sources: []Source{{
						Resolution: 1080,
						Url:        file.Url,
					}},
				})
			}
		}
	}

	return streams
}

func parseResolutionFromLabel(label string) (int, error) {
	match := rgxResolution.FindStringSubmatch(label)
	if len(match) < 2 {
		return 0, fmt.Errorf("no resolution height found in label")
	}
	res, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("atoi: %w", err)
	}
	return res, nil
}

func getSources(ctx context.Context, sps gql.StreamsParts, format string, defaultSourceLabel string, sortResolutionAsc bool) []Source {
	sourceMap := make(map[int]Source)

	for _, s := range sps.SceneStreams {
		if strings.Contains(s.Label, format) {
			resolution, err := parseResolutionFromLabel(s.Label)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("label", s.Label).Msg("Failed to parse resolution from label")
				continue
			}

			if _, ok := sourceMap[resolution]; ok {
				continue
			}

			sourceMap[resolution] = Source{
				Resolution: resolution,
				Url:        s.Url,
			}
		} else if s.Label == defaultSourceLabel {
			sourceMap[sps.Files[0].Height] = Source{
				Resolution: sps.Files[0].Height,
				Url:        s.Url,
			}
		}
	}
	sources := make([]Source, 0, len(sourceMap))
	for _, v := range sourceMap {
		sources = append(sources, v)
	}
	sortSourcesByResolution(sources, sortResolutionAsc)
	return sources
}

func sortSourcesByResolution(sources []Source, asc bool) {
	if asc {
		sort.Slice(sources, func(i, j int) bool { return sources[i].Resolution < sources[j].Resolution })
	} else {
		sort.Slice(sources, func(i, j int) bool { return sources[i].Resolution > sources[j].Resolution })
	}
}
