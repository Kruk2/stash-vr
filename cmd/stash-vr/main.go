//go:generate go run github.com/Khan/genqlient ../../internal/stash/gql/genqlient.yaml

package main

import (
	"stash-vr/cmd/stash-vr/internal"
	_ "stash-vr/internal/logger"

	"github.com/rs/zerolog/log"
)

func main() {
	if err := internal.Run(); err != nil {
		log.Warn().Err(err).Msg("Application EXIT with ERROR")
	} else {
		log.Info().Msg("Application EXIT without error")
	}
}
