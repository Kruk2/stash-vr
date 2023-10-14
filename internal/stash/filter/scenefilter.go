package filter

import (
	"fmt"
	"stash-vr/internal/stash/gql"
)

type Filter struct {
	FilterOpts  gql.FindFilterType
	SceneFilter gql.SceneFilterType
}

func SavedFilterToSceneFilter(savedFilter gql.SavedFilterParts) (Filter, error) {
	if savedFilter.Mode != gql.FilterModeScenes {
		return Filter{}, fmt.Errorf("unsupported filter mode")
	}

	filterQuery, err := parseJsonEncodedFilter(savedFilter)
	if err != nil {
		return Filter{}, fmt.Errorf("parseJsonEncodedFilter: %w", err)
	}
	return filterQuery, nil
}

func parseJsonEncodedFilter(stashFilter gql.SavedFilterParts) (Filter, error) {

	f, err := parseSceneFilterCriteria(stashFilter.Object_filter)
	if err != nil {
		return Filter{}, fmt.Errorf("parseSceneFilterCriteria: %w", err)
	}

	return Filter{FilterOpts: gql.FindFilterType{
		Per_page:  stashFilter.Find_filter.Per_page,
		Sort:      stashFilter.Find_filter.Sort,
		Direction: stashFilter.Find_filter.Direction,
	}, SceneFilter: f}, nil
}

func parseSceneFilterCriteria(jsonCriteria map[string]interface{}) (gql.SceneFilterType, error) {
	f := gql.SceneFilterType{}
	for name := range jsonCriteria {
		err := setSceneFilterCriterion(name, jsonCriteria[name].(map[string]interface{}), &f)
		if err != nil {
			return gql.SceneFilterType{}, fmt.Errorf("setSceneFilterCriterion: %w", err)
		}
	}
	return f, nil
}

func setSceneFilterCriterion(name string, criterionRaw map[string]interface{}, sceneFilter *gql.SceneFilterType) error {
	var err error
	criterion := jsonCriterion{Modifier: criterionRaw["modifier"].(string), Value: criterionRaw["value"]}
	switch name {
	//HierarchicalMultiCriterionInput
	case "tags":
		sceneFilter.Tags, err = criterion.asHierarchicalMultiCriterionInput()
		if err != nil {
			return fmt.Errorf("AsHierarchicalMultiCriterionInput: %w", err)
		}
	case "studios":
		sceneFilter.Studios, err = criterion.asHierarchicalMultiCriterionInput()
		if err != nil {
			return fmt.Errorf("AsHierarchicalMultiCriterionInput: %w", err)
		}
	case "performerTags":
		sceneFilter.Performer_tags, err = criterion.asHierarchicalMultiCriterionInput()
		if err != nil {
			return fmt.Errorf("AsHierarchicalMultiCriterionInput: %w", err)
		}

	//StringCriterionInput
	case "title":
		sceneFilter.Title, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "details":
		sceneFilter.Details, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "oshash":
		sceneFilter.Oshash, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "sceneChecksum":
		sceneFilter.Checksum, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "phash":
		sceneFilter.Phash, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "path":
		sceneFilter.Path, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "stash_id":
		sceneFilter.Stash_id, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "url":
		sceneFilter.Url, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "captions":
		sceneFilter.Captions, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	case "director":
		sceneFilter.Director, err = criterion.asStringCriterionInput()
		if err != nil {
			return fmt.Errorf("AsStringCriterionInput: %w", err)
		}
	//IntCriterionInput
	case "rating", "rating100":
		sceneFilter.Rating, err = criterion.asIntCriterionInput()
		if err != nil {
			return fmt.Errorf("AsIntCriterionInput: %w", err)
		}
	case "o_counter":
		sceneFilter.O_counter, err = criterion.asIntCriterionInput()
		if err != nil {
			return fmt.Errorf("AsIntCriterionInput: %w", err)
		}
	case "duration":
		sceneFilter.Duration, err = criterion.asIntCriterionInput()
		if err != nil {
			return fmt.Errorf("AsIntCriterionInput: %w", err)
		}
	case "tag_count":
		sceneFilter.Tag_count, err = criterion.asIntCriterionInput()
		if err != nil {
			return fmt.Errorf("AsIntCriterionInput: %w", err)
		}
	case "performer_age":
		sceneFilter.Performer_age, err = criterion.asIntCriterionInput()
		if err != nil {
			return fmt.Errorf("AsIntCriterionInput: %w", err)
		}
	case "performer_count":
		sceneFilter.Performer_count, err = criterion.asIntCriterionInput()
		if err != nil {
			return fmt.Errorf("AsIntCriterionInput: %w", err)
		}
	case "interactive_speed":
		sceneFilter.Interactive_speed, err = criterion.asIntCriterionInput()
		if err != nil {
			return fmt.Errorf("AsIntCriterionInput: %w", err)
		}
	case "file_count":
		sceneFilter.File_count, err = criterion.asIntCriterionInput()
		if err != nil {
			return fmt.Errorf("AsIntCriterionInput: %w", err)
		}
	//bool
	case "organized":
		sceneFilter.Organized, err = criterion.asBool()
		if err != nil {
			return fmt.Errorf("AsBool: %w", err)
		}
	case "performer_favorite":
		sceneFilter.Performer_favorite, err = criterion.asBool()
		if err != nil {
			return fmt.Errorf("AsBool: %w", err)
		}
	case "interactive":
		sceneFilter.Interactive, err = criterion.asBool()
		if err != nil {
			return fmt.Errorf("AsBool: %w", err)
		}

	//PHashDuplicationCriterionInput
	case "duplicated":
		sceneFilter.Duplicated, err = criterion.asPHashDuplicationCriterionInput()
		if err != nil {
			return fmt.Errorf("AsPHashDuplicationCriterionInput: %w", err)
		}

	//ResolutionCriterionInput
	case "resolution":
		sceneFilter.Resolution, err = criterion.asResolutionCriterionInput()
		if err != nil {
			return fmt.Errorf("AsResolutionCriterionInput: %w", err)
		}

	//string
	case "hasMarkers":
		sceneFilter.Has_markers, err = criterion.asString()
		if err != nil {
			return fmt.Errorf("AsString: %w", err)
		}
	case "sceneIsMissing":
		sceneFilter.Is_missing, err = criterion.asString()
		if err != nil {
			return fmt.Errorf("AsString: %w", err)
		}

	//MultiCriterionInput
	case "movies":
		sceneFilter.Movies, err = criterion.asMultiCriterionInput()
		if err != nil {
			return fmt.Errorf("AsMultiCriterionInput: %w", err)
		}
	case "performers":
		sceneFilter.Performers, err = criterion.asMultiCriterionInput()
		if err != nil {
			return fmt.Errorf("AsMultiCriterionInput: %w", err)
		}
	default:
		return fmt.Errorf("Unable to parse: %s", name)
	}

	return nil
}
