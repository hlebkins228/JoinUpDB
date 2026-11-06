package mongo

import (
	"JoinUp/internal/models"
	"time"
)

// mongoEvent is the MongoDB representation of models.Event.
type mongoEvent struct {
	ID              int                `bson:"_id"`
	CreatorID       int                `bson:"creator_id"`
	Name            string             `bson:"name"`
	Desc            string             `bson:"description"`
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
	EventTime       time.Time          `bson:"event_date"`
	TelegramChatURL string             `bson:"telegram_chat_url"`
	City            string             `bson:"city"`
	Creator         mongoEventMember   `bson:"creator"`
	Members         []mongoEventMember `bson:"members"`
	Categories      []mongoCategory    `bson:"categories"`
	Location        mongoLocation      `bson:"location"`
	ImageID         *int               `bson:"image_id,omitempty"`
	Deleted         bool               `bson:"deleted"`
}

type mongoEventMember struct {
	ID       int    `bson:"id"`
	Name     string `bson:"name"`
	AvatarID *int   `bson:"avatar_id,omitempty"`
}

type mongoUser struct {
	ID                    int       `bson:"_id"`
	Name                  string    `bson:"name"`
	Age                   int       `bson:"age"`
	Login                 string    `bson:"login"`
	Password              string    `bson:"password"`
	CreatedAt             time.Time `bson:"created_at"`
	City                  string    `bson:"city"`
	TelegramLogin         *string   `bson:"telegram_login,omitempty"`
	AvatarID              *int      `bson:"avatar_id,omitempty"`
	Role                  string    `bson:"role"`
	CreatedEventsCount    int       `bson:"-"`
	JoinedEventsCount     int       `bson:"-"`
	OrganizerRating       *float64  `bson:"organizer_rating,omitempty"`
	OrganizerReviewsCount int       `bson:"organizer_reviews_count"`
}

type mongoReview struct {
	ID        int               `bson:"_id"`
	EventID   int               `bson:"event_id"`
	AuthorID  int               `bson:"author_id"`
	Rating    int               `bson:"rating"`
	Text      string            `bson:"text"`
	CreatedAt time.Time         `bson:"created_at"`
	Author    mongoReviewAuthor `bson:"author"`
}

type mongoReviewAuthor struct {
	ID       int    `bson:"id"`
	Name     string `bson:"name"`
	AvatarID *int   `bson:"avatar_id,omitempty"`
}

type mongoImage struct {
	ID   int    `bson:"_id"`
	Name string `bson:"name"`
	Data []byte `bson:"data"`
}

type mongoCategory struct {
	ID   int    `bson:"_id"`
	Name string `bson:"name"`
}

type mongoLocation struct {
	Name      string  `bson:"name"`
	Longitude float64 `bson:"longitude"`
	Latitude  float64 `bson:"latitude"`
	Address   string  `bson:"address"`
}

// --- mongoEvent <-> models.Event ---

func (me *mongoEvent) toModel() *models.Event {
	return &models.Event{
		ID:              me.ID,
		CreatorID:       me.CreatorID,
		Name:            me.Name,
		Desc:            me.Desc,
		CreatedAt:       me.CreatedAt,
		UpdatedAt:       me.UpdatedAt,
		EventTime:       me.EventTime,
		TelegramChatURL: me.TelegramChatURL,
		City:            me.City,
		Creator:         me.Creator.toModel(),
		Members:         mongoMembersToModel(me.Members),
		Categories:      mongoCategoriesToModel(me.Categories),
		Location:        me.Location.toModel(),
		ImageID:         me.ImageID,
		Deleted:         me.Deleted,
	}
}

func fromModelEvent(e *models.Event) *mongoEvent {
	return &mongoEvent{
		ID:              e.ID,
		CreatorID:       e.CreatorID,
		Name:            e.Name,
		Desc:            e.Desc,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
		EventTime:       e.EventTime,
		TelegramChatURL: e.TelegramChatURL,
		City:            e.City,
		Creator:         fromModelEventMember(e.Creator),
		Members:         modelMembersToMongo(e.Members),
		Categories:      modelCategoriesToMongo(e.Categories),
		Location:        fromModelLocation(e.Location),
		ImageID:         e.ImageID,
		Deleted:         e.Deleted,
	}
}

// --- mongoEventMember <-> models.EventMember ---

func (m mongoEventMember) toModel() models.EventMember {
	return models.EventMember{
		ID:       m.ID,
		Name:     m.Name,
		AvatarID: m.AvatarID,
	}
}

func fromModelEventMember(m models.EventMember) mongoEventMember {
	return mongoEventMember{
		ID:       m.ID,
		Name:     m.Name,
		AvatarID: m.AvatarID,
	}
}

func mongoMembersToModel(mm []mongoEventMember) []models.EventMember {
	out := make([]models.EventMember, 0, len(mm))
	for _, m := range mm {
		out = append(out, m.toModel())
	}
	return out
}

func modelMembersToMongo(mm []models.EventMember) []mongoEventMember {
	out := make([]mongoEventMember, 0, len(mm))
	for _, m := range mm {
		out = append(out, fromModelEventMember(m))
	}
	return out
}

// --- mongoCategory <-> models.Category ---

func (mc *mongoCategory) toModel() models.Category {
	return models.Category{
		ID:   mc.ID,
		Name: mc.Name,
	}
}

func fromModelCategory(c models.Category) mongoCategory {
	return mongoCategory{
		ID:   c.ID,
		Name: c.Name,
	}
}

func mongoCategoriesToModel(mc []mongoCategory) []models.Category {
	out := make([]models.Category, 0, len(mc))
	for _, c := range mc {
		out = append(out, c.toModel())
	}
	return out
}

func modelCategoriesToMongo(cc []models.Category) []mongoCategory {
	out := make([]mongoCategory, 0, len(cc))
	for _, c := range cc {
		out = append(out, fromModelCategory(c))
	}
	return out
}

// --- mongoLocation <-> models.Location ---

func (ml *mongoLocation) toModel() models.Location {
	return models.Location{
		Name:      ml.Name,
		Longitude: ml.Longitude,
		Latitude:  ml.Latitude,
		Address:   ml.Address,
	}
}

func fromModelLocation(l models.Location) mongoLocation {
	return mongoLocation{
		Name:      l.Name,
		Longitude: l.Longitude,
		Latitude:  l.Latitude,
		Address:   l.Address,
	}
}

// --- mongoUser <-> models.User ---

func (mu *mongoUser) toModel() *models.User {
	return &models.User{
		ID:                    mu.ID,
		Name:                  mu.Name,
		Age:                   mu.Age,
		Login:                 mu.Login,
		Password:              mu.Password,
		CreatedAt:             mu.CreatedAt,
		City:                  mu.City,
		TelegramLogin:         mu.TelegramLogin,
		AvatarID:              mu.AvatarID,
		Role:                  mu.Role,
		CreatedEventsCount:    mu.CreatedEventsCount,
		JoinedEventsCount:     mu.JoinedEventsCount,
		OrganizerRating:       mu.OrganizerRating,
		OrganizerReviewsCount: mu.OrganizerReviewsCount,
	}
}

func fromModelUser(u *models.User) *mongoUser {
	return &mongoUser{
		ID:                    u.ID,
		Name:                  u.Name,
		Age:                   u.Age,
		Login:                 u.Login,
		Password:              u.Password,
		CreatedAt:             u.CreatedAt,
		City:                  u.City,
		TelegramLogin:         u.TelegramLogin,
		AvatarID:              u.AvatarID,
		Role:                  u.Role,
		CreatedEventsCount:    u.CreatedEventsCount,
		JoinedEventsCount:     u.JoinedEventsCount,
		OrganizerRating:       u.OrganizerRating,
		OrganizerReviewsCount: u.OrganizerReviewsCount,
	}
}

// --- mongoReview <-> models.Review ---

func (mr *mongoReview) toModel() *models.Review {
	return &models.Review{
		ID:        mr.ID,
		EventID:   mr.EventID,
		AuthorID:  mr.AuthorID,
		Rating:    mr.Rating,
		Text:      mr.Text,
		CreatedAt: mr.CreatedAt,
		Author:    mr.Author.toModel(),
	}
}

func fromModelReview(r *models.Review) *mongoReview {
	return &mongoReview{
		ID:        r.ID,
		EventID:   r.EventID,
		AuthorID:  r.AuthorID,
		Rating:    r.Rating,
		Text:      r.Text,
		CreatedAt: r.CreatedAt,
		Author:    fromModelReviewAuthor(r.Author),
	}
}

// --- mongoReviewAuthor <-> models.ReviewAuthor ---

func (ma mongoReviewAuthor) toModel() models.ReviewAuthor {
	return models.ReviewAuthor{
		ID:       ma.ID,
		Name:     ma.Name,
		AvatarID: ma.AvatarID,
	}
}

func fromModelReviewAuthor(a models.ReviewAuthor) mongoReviewAuthor {
	return mongoReviewAuthor{
		ID:       a.ID,
		Name:     a.Name,
		AvatarID: a.AvatarID,
	}
}

// --- mongoImage <-> models.Image ---

func (mi *mongoImage) toModel() *models.Image {
	return &models.Image{
		Name: mi.Name,
		Data: mi.Data,
	}
}

func fromModelImage(img *models.Image) *mongoImage {
	return &mongoImage{
		Name: img.Name,
		Data: img.Data,
	}
}
