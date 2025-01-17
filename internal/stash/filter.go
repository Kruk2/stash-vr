package stash

import (
	"context"
	"fmt"
	"stash-vr/internal/stash/gql"
	"strconv"

	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
)

func FindFiltersByName(ctx context.Context, client graphql.Client, filterNames []string) []gql.SavedFilterParts {
	filters := make([]gql.SavedFilterParts, 0, len(filterNames))
	response, _ := gql.FindSavedSceneFilters(ctx, client)

	for _, filterName := range filterNames {
		found := false
		for _, filter := range response.FindSavedFilters {
			if filter.Name == filterName {
				filters = append(filters, filter.SavedFilterParts)
				found = true
				break
			}
		}

		if !found {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("FindFiltersById: FindSavedFilter: Filter not found")).Str("filterName", filterName).Msg("Skipped filter")
			continue
		}
	}

	return filters
}

func FindFiltersById(ctx context.Context, client graphql.Client, filterIds []string) []gql.SavedFilterParts {
	filters := make([]gql.SavedFilterParts, 0, len(filterIds))

	for _, filterId := range filterIds {
		savedFilterResponse, err := gql.FindSavedFilter(ctx, client, filterId)
		if err != nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("FindFiltersById: FindSavedFilter: %w", err)).Str("filterId", filterId).Msg("Skipped filter")
			continue
		}
		if savedFilterResponse.FindSavedFilter == nil {
			log.Ctx(ctx).Warn().Err(fmt.Errorf("FindFiltersById: FindSavedFilter: Filter not found")).Str("filterId", filterId).Msg("Skipped filter")
			continue
		}
		filters = append(filters, savedFilterResponse.FindSavedFilter.SavedFilterParts)
	}

	return filters
}

func FindSavedFilterIdsByFrontPage(ctx context.Context, client graphql.Client) ([]string, error) {
	configurationResponse, err := gql.UIConfiguration(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("UIConfiguration: %w", err)
	}

	frontPageContent := configurationResponse.Configuration.Ui["frontPageContent"]
	if frontPageContent == nil {
		log.Ctx(ctx).Info().Msg("No frontpage content found")
		return nil, nil
	}

	frontPageFilters := configurationResponse.Configuration.Ui["frontPageContent"].([]interface{})
	filterIds := make([]string, 0, len(frontPageFilters))
	for _, _filter := range frontPageFilters {
		filter := _filter.(map[string]interface{})
		typeName := filter["__typename"].(string)
		if typeName == "CustomFilter" {
			log.Ctx(ctx).Debug().Msg("Filter skipped: Predefined filter on front page: Only user created saved scene filters are supported.")
			continue
		}
		if typeName != "SavedFilter" {
			log.Ctx(ctx).Debug().Str("type", typeName).Msg("Filter skipped: Filter of unsupported type on front page: Only user created saved scene filters are supported")
			continue
		}

		filterId := strconv.Itoa(int(filter["savedFilterId"].(float64)))
		filterIds = append(filterIds, filterId)
	}

	return filterIds, nil
}
