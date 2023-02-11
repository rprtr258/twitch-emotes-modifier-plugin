package seventv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rprtr258/xerr"
)

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	CreatedAt   int64  `json:"created_at"`
	AvatarURL   string `json:"avatar_url"`
	EmoteSets   []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Flags    int    `json:"flags"`
		Capacity int    `json:"capacity"`
	} `json:"emote_sets"`
	Editors []struct {
		ID          string `json:"id"`
		Permissions int    `json:"permissions"`
		Visible     bool   `json:"visible"`
		AddedAt     int64  `json:"added_at"`
	} `json:"editors"`
	Roles       []string `json:"roles"`
	Connections []struct {
		ID            string      `json:"id"`
		Platform      string      `json:"platform"`
		Username      string      `json:"username"`
		DisplayName   string      `json:"display_name"`
		LinkedAt      int64       `json:"linked_at"`
		EmoteCapacity int         `json:"emote_capacity"`
		EmoteSetID    interface{} `json:"emote_set_id"`
		EmoteSet      struct {
			ID         string        `json:"id"`
			Name       string        `json:"name"`
			Flags      int           `json:"flags"`
			Tags       []interface{} `json:"tags"`
			Immutable  bool          `json:"immutable"`
			Privileged bool          `json:"privileged"`
			Capacity   int           `json:"capacity"`
			Owner      interface{}   `json:"owner"`
		} `json:"emote_set"`
	} `json:"connections"`
}

type EmoteSet struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Flags      int           `json:"flags"`
	Tags       []interface{} `json:"tags"`
	Immutable  bool          `json:"immutable"`
	Privileged bool          `json:"privileged"`
	Emotes     []struct {
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		Flags     int         `json:"flags"`
		Timestamp int64       `json:"timestamp"`
		ActorID   interface{} `json:"actor_id"`
		Data      struct {
			ID        string   `json:"id"`
			Name      string   `json:"name"`
			Flags     int      `json:"flags"`
			Lifecycle int      `json:"lifecycle"`
			State     []string `json:"state"`
			Listed    bool     `json:"listed"`
			Animated  bool     `json:"animated"`
			Owner     struct {
				ID          string `json:"id"`
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
				AvatarURL   string `json:"avatar_url"`
				Style       struct {
				} `json:"style"`
				Roles []string `json:"roles"`
			} `json:"owner"`
			Host struct {
				URL   string `json:"url"`
				Files []struct {
					Name       string `json:"name"`
					StaticName string `json:"static_name"`
					Width      int    `json:"width"`
					Height     int    `json:"height"`
					FrameCount int    `json:"frame_count"`
					Size       int    `json:"size"`
					Format     string `json:"format"`
				} `json:"files"`
			} `json:"host"`
		} `json:"data,omitempty"`
	} `json:"emotes"`
	EmoteCount int `json:"emote_count"`
	Capacity   int `json:"capacity"`
	Owner      struct {
		ID          string `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Style       struct {
		} `json:"style"`
		Roles []string `json:"roles"`
	} `json:"owner"`
}

type Repository struct {
}

func (Repository) Download7tvEmote(emoteID, outObjectID string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("https://cdn.7tv.app/emote/%s/4x", emoteID))
	if err != nil {
		return nil, xerr.New(
			xerr.WithErr(err),
			xerr.WithMessage("failed downloading emote"),
			xerr.WithField("emoteID", emoteID),
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, xerr.New(
			xerr.WithMessage("bad status for emote request"),
			xerr.WithField("statusCode", resp.StatusCode),
		)
	}

	// var extension string
	switch imageFormat := resp.Header.Get("Content-type"); imageFormat {
	case "image/webp":
		// extension = "webp"
	default:
		return nil, xerr.New(
			xerr.WithMessage("unknown image format"),
			xerr.WithField("imageFormat", imageFormat),
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (Repository) GetUser(userID string) (User, error) {
	userResp, err := http.Get(fmt.Sprintf("https://7tv.io/v3/users/%s", userID))
	if err != nil {
		return User{}, xerr.New(
			xerr.WithErr(err),
			xerr.WithMessage("failed getting 7tv user"),
			xerr.WithField("userID", userID),
		)
	}
	defer userResp.Body.Close()

	var user User
	if err := json.NewDecoder(userResp.Body).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

func (r Repository) GetEmoteSet(userID string) (EmoteSet, error) {
	userResp, err := http.Get(fmt.Sprintf("https://7tv.io/v3/emote-sets/%s", userID))
	if err != nil {
		return EmoteSet{}, err
	}
	defer userResp.Body.Close()

	var emoteSet EmoteSet
	if err := json.NewDecoder(userResp.Body).Decode(&emoteSet); err != nil {
		return EmoteSet{}, err
	}

	return emoteSet, nil
}

func (r Repository) GetEmoteID(userID string, emoteName string) (string, error) {
	user, err := r.GetUser(userID)
	if err != nil {
		return "", xerr.New(
			xerr.WithMessage("failed getting user, while getting emote id"),
			xerr.WithErr(err),
			xerr.WithField("emoteName", emoteName),
		)
	}

	var emoteSetID string
	for _, connection := range user.Connections {
		if connection.Platform == "TWITCH" {
			emoteSetID = connection.EmoteSet.ID
		}
	}
	if emoteSetID == "" {
		return "", xerr.New(xerr.WithMessage("no emote set found for twitch"))
	}

	emoteSet, err := r.GetEmoteSet(emoteSetID)
	if err != nil {
		return "", err
	}

	for _, emote := range emoteSet.Emotes {
		if emote.Name == emoteName {
			return emote.ID, nil
		}
	}

	return "", xerr.New(
		xerr.WithMessage("no emote found"),
		xerr.WithField("userID", userID),
		xerr.WithField("emoteName", emoteName),
	)
}
