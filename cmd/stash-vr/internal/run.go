package internal

import (
	"context"
	"fmt"
	"stash-vr/internal/application"
	"stash-vr/internal/config"
	"stash-vr/internal/sections"
	"stash-vr/internal/server"
	"stash-vr/internal/stash"
	"stash-vr/internal/stash/gql"

	"github.com/Khan/genqlient/graphql"
	"github.com/rs/zerolog/log"
)

const listenAddress = "0.0.0.0:9664"

func Run() error {
	ctx := application.InterruptableContext()

	log.Info().Str("config", fmt.Sprintf("%+v", config.Get().Redacted())).Send()

	stashClient := stash.NewClient(config.Get().StashGraphQLUrl, config.Get().StashApiKey)

	logVersions(ctx, stashClient)

	sections.Get(ctx, stashClient)

	err := server.Listen(ctx, listenAddress, stashClient)
	if err != nil {
		return fmt.Errorf("server: %w", err)
	}

	return nil
}

func logVersions(ctx context.Context, client graphql.Client) {
	log.Info().Str("Stash-VR version", application.BuildVersion).Send()

	if version, err := gql.Version(ctx, client); err != nil {
		log.Warn().Err(err).Msg("Failed to retrieve stash version")
	} else {
		log.Info().Str("Stash version", version.Version.Version).Send()
	}
}
