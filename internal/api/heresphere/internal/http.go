package internal

import (
	"encoding/json"
	"github.com/Khan/genqlient/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"stash-vr/internal/api/heresphere/internal/index"
	"stash-vr/internal/api/heresphere/internal/scan"
	"stash-vr/internal/api/heresphere/internal/sync"
	"stash-vr/internal/api/heresphere/internal/videodata"
	"stash-vr/internal/api/internal"
	"stash-vr/internal/util"
)

type HttpHandler struct {
	Client graphql.Client
}

func (h *HttpHandler) Index(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := util.GetBaseUrl(req)

	data := index.Build(ctx, h.Client, baseUrl)

	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *HttpHandler) Scan(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	baseUrl := util.GetBaseUrl(req)

	data, err := scan.Build(ctx, h.Client, baseUrl)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("scan")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}
}

func (h *HttpHandler) VideoData(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	ctx := req.Context()
	sceneId := chi.URLParam(req, "videoId")

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("body: read")
		return
	}

	var updateVideoData sync.UpdateVideoData
	err = json.Unmarshal(body, &updateVideoData)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Bytes("body", body).Msg("body: unmarshal")
	} else {
		if updateVideoData.IsUpdateRequest() {
			sync.Update(ctx, h.Client, sceneId, updateVideoData)
			w.WriteHeader(http.StatusOK)
			return
		}

		if updateVideoData.IsDeleteRequest() {
			sync.Destroy(ctx, h.Client, sceneId)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	data, err := videodata.Build(ctx, h.Client, sceneId)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("build")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := internal.WriteJson(ctx, w, data); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("write")
	}

}
