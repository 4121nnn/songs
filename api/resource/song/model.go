package song

import (
	"github.com/google/uuid"
)

type DTO struct {
	ReleaseDate string `json:"release_date"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type Form struct {
	Group string `json:"group" form:"required"`
	Song  string `json:"song" form:"required"`
}

type Song struct {
	ID          uuid.UUID `gorm:"primarykey" json:"id"`
	Group       string    `gorm:"column:group_name" json:"group"`
	Song        string    `gorm:"column:song_name" json:"song"`
	Text        string    `gorm:"column:text" json:"text"`
	ReleaseDate string    `gorm:"column:release_date" json:"release_date"`
	Link        string    `gorm:"column:link" json:"link"`
}

type SongRequest struct {
	Group       string `json:"group" binding:"required"`
	Song        string `json:"song" binding:"required"`
	Text        string `json:"text" binding:"required"`
	ReleaseDate string `json:"release_date" binding:"required"`
	Link        string `json:"link" binding:"required"`
}

type Songs []*Song

func (s *Song) ToDto() *DTO {
	return &DTO{
		ReleaseDate: s.ReleaseDate,
		Text:        s.Text,
		Link:        s.Link,
	}
}

func (songs Songs) ToDto() []*DTO {
	dtos := make([]*DTO, len(songs))
	for i, v := range songs {
		dtos[i] = v.ToDto()
	}

	return dtos
}

func (f *Form) ToModel() *Song {

	return &Song{
		Group: f.Group,
		Song:  f.Song,
	}
}
