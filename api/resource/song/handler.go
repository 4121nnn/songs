package song

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"net/http"
	e "songs/api/resource/common/err"
	l "songs/api/resource/common/log"
	"songs/pkg/pagination"
	ctxUtil "songs/util/ctx"
	validatorUtil "songs/util/validator"
)

type API struct {
	logger     *zerolog.Logger
	validator  *validator.Validate
	repository *Repository
}

func New(logger *zerolog.Logger, validator *validator.Validate, db *gorm.DB) *API {
	return &API{
		logger:     logger,
		validator:  validator,
		repository: NewRepository(db, logger),
	}
}

// List godoc
//
//	@summary		List songs
//	@description	List songs with pagination and optional filters.
//	@tags			songs
//	@accept			json
//	@produce		json
//	@param			page		query		int					false	"Page number (default is 1)"
//	@param			pageSize	query		int					false	"Number of items per page (default is 10, max is 100)"
//	@param			group		query		string				false	"Group name"
//	@param			song		query		string				false	"Song name"
//	@param			text		query		string				false	"Text to search within song lyrics"
//	@param			releaseDate	query		string				false	"Release date"
//	@param			link		query		string				false	"Song link"
//	@success		200			{object}	pagination.Pages	"Paginated list of songs"
//	@failure		500			{object}	err.Error			"Internal server error"
//	@router			/ [get]
func (a *API) List(w http.ResponseWriter, r *http.Request) {
	reqID := ctxUtil.RequestID(r.Context())

	a.logger.Debug().Str(l.KeyReqID, reqID).Msg("List function started")

	// Get pagination parameters
	pages := pagination.NewFromRequest(r, -1) // Using -1 to indicate unknown total count
	a.logger.Debug().Str(l.KeyReqID, reqID).Int("page", pages.Page).Int("perPage", pages.PerPage).Msg("Pagination parameters retrieved")

	// Get filter parameters
	filters := map[string]interface{}{}
	if group := r.URL.Query().Get("group"); group != "" {
		filters["group_name"] = group
		a.logger.Debug().Str(l.KeyReqID, reqID).Str("group_name", group).Msg("Filter added: group_name")
	}
	if song := r.URL.Query().Get("song"); song != "" {
		filters["song_name"] = song
		a.logger.Debug().Str(l.KeyReqID, reqID).Str("song_name", song).Msg("Filter added: song_name")
	}
	if text := r.URL.Query().Get("text"); text != "" {
		filters["text"] = text
		a.logger.Debug().Str(l.KeyReqID, reqID).Str("text", text).Msg("Filter added: text")
	}
	if releaseDate := r.URL.Query().Get("releaseDate"); releaseDate != "" {
		filters["release_date"] = releaseDate
		a.logger.Debug().Str(l.KeyReqID, reqID).Str("release_date", releaseDate).Msg("Filter added: release_date")
	}
	if link := r.URL.Query().Get("link"); link != "" {
		filters["link"] = link
		a.logger.Debug().Str(l.KeyReqID, reqID).Str("link", link).Msg("Filter added: link")
	}

	// Call the repository's List method with pagination and filters
	page, err := a.repository.List(pages.Page, pages.PerPage, filters)
	if err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to retrieve paginated songs from repository")
		e.ServerError(w, e.RespDBDataAccessFailure)
		return
	}

	// Set the link header for pagination
	w.Header().Set("Link", page.BuildLinkHeader(r.URL.String(), pagination.DefaultPageSize))

	// Return the paginated response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(page); err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("")
		e.ServerError(w, e.RespJSONEncodeFailure)
		return
	}

	a.logger.Info().Str(l.KeyReqID, reqID).Msg("Paginated songs retrieved successfully")
}

// Create godoc
//
//	@summary		Create song
//	@description	Create song
//	@tags			songs
//	@accept			json
//	@produce		json
//	@success		201
//	@failure		400	{object}	err.Error
//	@failure		422	{object}	err.Errors
//	@failure		500	{object}	err.Error
//	@router			/ [post]
//	@param			song	body	SongRequest	true	"The song details for creation"
//
// The Song struct requires the following fields:
// - Group (string): Name of the group or artist (required).
// - Song (string): Title of the song (required).
// - Text (string): Lyrics or text of the song (required).
// - ReleaseDate (string): Release date of the song in YYYY-MM-DD format (required).
// - Link (string): URL link related to the song (required).
func (a *API) Create(w http.ResponseWriter, r *http.Request) {
	reqID := ctxUtil.RequestID(r.Context())

	a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Create function started")

	song := &Song{}

	if err := json.NewDecoder(r.Body).Decode(song); err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to decode JSON")
		e.BadRequest(w, e.RespJSONDecodeFailure)
		return
	}

	if err := a.validator.Struct(song); err != nil {
		respBody, err := json.Marshal(validatorUtil.ToErrResponse(err))
		if err != nil {
			a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to encode validation errors to JSON")
			e.ServerError(w, e.RespJSONEncodeFailure)
			return
		}

		a.logger.Debug().Str(l.KeyReqID, reqID).Msgf("Validation errors: %s", respBody)

		e.ValidationErrors(w, respBody)
		return
	}

	song.ID = uuid.New()

	a.logger.Debug().Str(l.KeyReqID, reqID).Msgf("Creating new song: %+v", song)

	song, err := a.repository.Create(song)
	if err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("")
		e.ServerError(w, e.RespDBDataInsertFailure)
		return
	}

	a.logger.Info().Str(l.KeyReqID, reqID).Str("id", song.ID.String()).Msg("New song created")
	w.WriteHeader(http.StatusCreated)
}

// Read godoc
//
//	@summary		Read song
//	@description	Read song
//	@tags			songs
//	@accept			json
//	@produce		json
//	@param			id	path		string	true	"Song ID"
//	@success		200	{object}	Song
//	@failure		400	{object}	err.Error
//	@failure		404
//	@failure		500	{object}	err.Error
//	@router			/{id} [get]
func (a *API) Read(w http.ResponseWriter, r *http.Request) {
	reqID := ctxUtil.RequestID(r.Context())

	a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Read function started")

	id, err := uuid.Parse(chi.URLParam(r, "id"))

	if err != nil {
		e.BadRequest(w, e.RespInvalidURLParamID)
		return
	}

	a.logger.Debug().Str(l.KeyReqID, reqID).Msgf("Parsed ID: %s", id.String())

	song, err := a.repository.Read(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Song not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to access the song in the database")
		e.ServerError(w, e.RespDBDataAccessFailure)
		return
	}

	a.logger.Debug().Str(l.KeyReqID, reqID).Msgf("Retrieved song: %+v", song)

	if err := json.NewEncoder(w).Encode(song); err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to encode song DTO to JSON")
		e.ServerError(w, e.RespJSONEncodeFailure)
		return
	}

	a.logger.Info().Str(l.KeyReqID, reqID).Msg("Response successfully encoded and sent")
}

//	 GetLyrics godoc
//
//		@summary		Get song lyrics
//		@description	Get lyrics for a specific song and group
//		@tags			songs
//		@accept			json
//		@produce		json
//		@param			group	query		string	true	"Group name"
//		@param			song	query		string	true	"Song name"
//		@success		200		{object}	DTO     "Successfully retrieved the song lyrics"
//		@failure		400		{object}	err.Error
//		@failure		404
//		@failure		500	{object}	err.Error
//		@router			/info [get]
func (a *API) Info(w http.ResponseWriter, r *http.Request) {
	reqID := ctxUtil.RequestID(r.Context())

	a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Info function started")

	// Extract the 'group' and 'song' from the URL query parameters
	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")

	// Validate song parameters (e.g., ensure they are not empty)
	if group == "" || song == "" {
		a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Invalid query parameters: group or song is empty")
		e.BadRequest(w, e.RespInvalidURLParamID)
		return
	}

	pages := pagination.NewFromRequest(r, -1) // Using -1 to indicate unknown total count

	pages.PerPage = 4
	// Fetch the song dto based on the group and song name
	dto, err := a.repository.GetLyrics(group, song, pages.Page, pages.PerPage)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Song not found in the repository")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to fetch song")
		e.ServerError(w, e.RespDBDataAccessFailure)
		return
	}

	a.logger.Debug().Str(l.KeyReqID, reqID).Msgf("Retrieved DTO: %+v", dto)

	// Return the dto as JSON response
	if err := json.NewEncoder(w).Encode(dto); err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to encode dto to JSON")
		e.ServerError(w, e.RespJSONEncodeFailure)
		return
	}

	a.logger.Info().Str(l.KeyReqID, reqID).Msg("Response successfully encoded and sent")
}

// Update godoc
//
//	@summary		Update song
//	@description	Update song
//	@tags			songs
//	@accept			json
//	@produce		json
//	@param			id		path	string	true	"Song ID"
//	@param			body	body	Form	true	"Song form"
//	@success		200
//	@failure		400	{object}	err.Error
//	@failure		404
//	@failure		422	{object}	err.Errors
//	@failure		500	{object}	err.Error
//	@router			/{id} [put]
func (a *API) Update(w http.ResponseWriter, r *http.Request) {
	reqID := ctxUtil.RequestID(r.Context())

	a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Update function started")

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Invalid UUID in URL parameter")
		e.BadRequest(w, e.RespInvalidURLParamID)
		return
	}

	a.logger.Debug().Str(l.KeyReqID, reqID).Str("id", id.String()).Msg("Parsed ID from URL parameter")

	song := &Song{}
	if err := json.NewDecoder(r.Body).Decode(song); err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to decode JSON request body")
		e.BadRequest(w, e.RespJSONDecodeFailure)
		return
	}

	a.logger.Debug().Str(l.KeyReqID, reqID).Msgf("Decoded song: %+v", song)

	if err := a.validator.Struct(song); err != nil {
		respBody, err := json.Marshal(validatorUtil.ToErrResponse(err))
		if err != nil {
			a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to encode validation error response to JSON")
			e.ServerError(w, e.RespJSONEncodeFailure)
			return
		}

		a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Validation errors occurred")
		e.ValidationErrors(w, respBody)
		return
	}

	song.ID = id

	a.logger.Debug().Str(l.KeyReqID, reqID).Msgf("Updating song: %+v", song)

	rows, err := a.repository.Update(song)
	if err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to update song in the repository")
		e.ServerError(w, e.RespDBDataUpdateFailure)
		return
	}
	if rows == 0 {
		a.logger.Debug().Str(l.KeyReqID, reqID).Msg("No rows affected; song not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	a.logger.Info().Str(l.KeyReqID, reqID).Str("id", id.String()).Msg("Song updated successfully")
}

// Delete godoc
//
//	@summary		Delete song
//	@description	Delete song
//	@tags			songs
//	@accept			json
//	@produce		json
//	@param			id	path	string	true	"Song ID"
//	@success		200
//	@failure		400	{object}	err.Error
//	@failure		404
//	@failure		500	{object}	err.Error
//	@router			/{id} [delete]
func (a *API) Delete(w http.ResponseWriter, r *http.Request) {
	reqID := ctxUtil.RequestID(r.Context())

	a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Delete function started")

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		a.logger.Debug().Str(l.KeyReqID, reqID).Msg("Invalid UUID in URL parameter")
		e.BadRequest(w, e.RespInvalidURLParamID)
		return
	}

	a.logger.Debug().Str(l.KeyReqID, reqID).Str("id", id.String()).Msg("Parsed ID from URL parameter")

	rows, err := a.repository.Delete(id)
	if err != nil {
		a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("Failed to delete song from the repository")
		e.ServerError(w, e.RespDBDataRemoveFailure)
		return
	}
	if rows == 0 {
		a.logger.Debug().Str(l.KeyReqID, reqID).Msg("No rows affected; song not found for deletion")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	a.logger.Info().Str(l.KeyReqID, reqID).Str("id", id.String()).Msg("Song deleted successfully")
}
