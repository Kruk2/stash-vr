package sections

import (
	"bufio"
	"context"
	"os"
	"stash-vr/internal/cache"
	"stash-vr/internal/config"
	"stash-vr/internal/sections/internal"
	"stash-vr/internal/sections/section"

	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
)

var c cache.Cache[[]section.Section]

func Get(ctx context.Context, client graphql.Client) []section.Section {
	return c.Get(ctx, func(ctx context.Context) []section.Section {
		return build(ctx, client, config.Get().Filters)
	})
}

func build(ctx context.Context, client graphql.Client, filters string) []section.Section {
	sss := make([][]section.Section, 3)

	readFile, err := os.Open("sections.txt")

	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by filter ids")
		return nil
	}

	filterNames := []string{}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		filterName := fileScanner.Text()
		if filterName != "" {
			filterNames = append(filterNames, filterName)
		}
	}
	readFile.Close()

	ss, err := internal.SectionsByFilterName(ctx, client, "", filterNames)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to build sections by filter ids")
		return nil
	}
	sss[2] = ss
	log.Ctx(ctx).Debug().Int("count", len(ss)).Msg("Sections built from filter list")

	var sections []section.Section

	for _, ss := range sss {
		for _, s := range ss {
			if s.FilterId != "" && section.ContainsFilterId(s.FilterId, sections) {
				log.Ctx(ctx).Trace().Str("filterId", s.FilterId).Str("section", s.Name).Msg("Filter already added, skipping")
				continue
			}
			sections = append(sections, s)
		}
	}

	if len(sections) == 0 {
		log.Ctx(ctx).Info().Msg("No scenes found using current filters. Adding a default section with all scenes.")
		s, err := internal.SectionWithAllScenes(ctx, client)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to build custom section with all scenes")
		} else {
			if len(s.PreviewPartsList) == 0 {
				log.Ctx(ctx).Info().Msg("No scenes found in Stash.")
			} else {
				sections = append(sections, s)
			}
		}
	}

	count := Count(sections)

	if count.Links > 10000 {
		log.Ctx(ctx).Warn().Int("links", count.Links).Msg("More than 10.000 links generated. Known to cause issues with video players.")
	}

	log.Ctx(ctx).Info().Int("sections", len(sections)).Int("links", count.Links).Int("scenes", count.Scenes).Msg("Sections build complete")

	return sections
}
