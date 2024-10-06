package song

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"songs/pkg/pagination"
	"strings"
)

type Repository struct {
	db     *gorm.DB
	logger *zerolog.Logger
}

func NewRepository(db *gorm.DB, l *zerolog.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: l,
	}
}

func (r *Repository) List(page, pageSize int, filters map[string]interface{}) (*pagination.Pages, error) {
	var songs []Song
	var total int64

	r.logger.Debug().Msgf("List called with page: %d, pageSize: %d, filters: %+v", page, pageSize, filters)

	// Create a base query
	query := r.db.Model(&Song{})

	// Apply filters to the query
	for key, value := range filters {
		if key == "text" {
			query = query.Where("text LIKE ?", "%"+value.(string)+"%")

			r.logger.Debug().Msgf("Applying filter: %s LIKE %s", key, value)
		} else {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)

			r.logger.Debug().Msgf("Applying filter: %s = %v", key, value)
		}
	}

	// Get the total count of records matching the filters
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&songs).Error; err != nil {
		return nil, err
	}

	r.logger.Debug().Msgf("Retrieved %d songs for page: %d", len(songs), page)

	pages := pagination.New(page, pageSize, int(total))
	pages.Items = songs

	return pages, nil
}

func (r *Repository) GetLyrics(group, song string, page, pageSize int) (*pagination.Pages, error) {
	r.logger.Debug().Msgf("GetLyrics called with group: %s, song: %s, page: %d, pageSize: %d", group, song, page, pageSize)

	s := &Song{}

	if err := r.db.Where("group_name = ? AND song_name = ?", group, song).First(s).Error; err != nil {
		return nil, err
	}

	// Split the lyrics by newline character to get individual verses
	allVerses := strings.Split(s.Text, "\\n")
	total := len(allVerses)

	r.logger.Debug().Msgf("Total verses found: %d", total)

	// Calculate pagination limits
	start := (page - 1) * pageSize
	end := start + pageSize

	if end > total {
		end = total
	}

	r.logger.Debug().Msgf("Pagination limits - start: %d, end: %d", start, end)

	var paginatedLyrics string
	for i := start; i < end; i++ {
		paginatedLyrics += allVerses[i]
		paginatedLyrics += "\n"
	}
	s.Text = paginatedLyrics

	pages := pagination.New(page, pageSize, int(total))
	pages.Items = s

	r.logger.Debug().Msgf("Returning pages with total verses: %d, current page: %d", total, page)

	return pages, nil
}

func (r *Repository) Create(song *Song) (*Song, error) {
	r.logger.Debug().Msgf("Attempting to create a new song: %+v", song)

	if err := r.db.Create(song).Error; err != nil {
		return nil, err
	}

	r.logger.Debug().Msgf("Successfully created song with ID: %d", song.ID)
	return song, nil
}

func (r *Repository) Read(id uuid.UUID) (*Song, error) {
	song := &Song{}
	if err := r.db.Where("id = ?", id).First(&song).Error; err != nil {
		return nil, err
	}

	return song, nil
}

func (r *Repository) Update(song *Song) (int64, error) {
	r.logger.Debug().Msgf("Attempting to update song with ID: %d, data: %+v", song.ID, song)

	result := r.db.Model(&Song{}).
		Select("Group", "Song", "Text", "ReleaseDate", "Link").
		Where("id = ?", song.ID).
		Updates(song)

	r.logger.Debug().Msgf("Successfully updated song with ID: %d, rows affected: %d", song.ID, result.RowsAffected)
	return result.RowsAffected, result.Error
}

func (r *Repository) Delete(id uuid.UUID) (int64, error) {
	r.logger.Debug().Msgf("Attempting to delete song with ID: %s", id.String())

	result := r.db.Where("id = ?", id).Delete(&Song{})

	r.logger.Debug().Msgf("Successfully deleted song with ID: %s, rows affected: %d", id.String(), result.RowsAffected)
	return result.RowsAffected, result.Error
}
