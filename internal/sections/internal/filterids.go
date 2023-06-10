package internal

import (
	"context"
	"stash-vr/internal/sections/section"
	"stash-vr/internal/stash"

	"github.com/Khan/genqlient/graphql"
)

func SectionsByFilterName(ctx context.Context, client graphql.Client, prefix string, filterNames []string) ([]section.Section, error) {
	savedFilters := stash.FindFiltersByName(ctx, client, filterNames)

	sections := sectionFromSavedFilterFuncBuilder(ctx, client, prefix, "Filter List").Ordered(savedFilters)

	return sections, nil
}

func SectionsByFilterIds(ctx context.Context, client graphql.Client, prefix string, filterIds []string) ([]section.Section, error) {
	savedFilters := stash.FindFiltersById(ctx, client, filterIds)

	sections := sectionFromSavedFilterFuncBuilder(ctx, client, prefix, "Filter List").Ordered(savedFilters)

	return sections, nil
}
