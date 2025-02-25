package dousetsu

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/machinebox/graphql"
)

type UserResponse struct {
	User struct {
		ID              string `json:"id"`
		Login           string `json:"login"`
		DisplayName     string `json:"displayName"`
		ProfileImageURL string `json:"profileImageURL"`
		Followers       struct {
			TotalCount int `json:"totalCount"`
		} `json:"followers"`
		Stream struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			Type         string `json:"type"`
			ViewersCount int    `json:"viewersCount"`
			CreatedAt    string `json:"createdAt"`
			Game         struct {
				Name string `json:"name"`
			} `json:"game"`
		} `json:"stream"`
	} `json:"user"`
}

const query = `
	query GetUser($login: String!) {
		user(login: $login) {
			id
			login
			displayName
			profileImageURL(width: 300)
			followers {
				totalCount
			}
			stream {
				id
				title
				type
				viewersCount
				createdAt
				game {
					name
				}
			}
		}
	}`

const (
	twitchGQLEndpoint = "https://gql.twitch.tv/gql"
	twitchWSURL       = "wss://pubsub-edge.twitch.tv/v1"
	twitchClientID    = "kimne78kx3ncx6brgo4mv6wki5h1ko"
	twitchPubSubNonce = "random-string"
)

type TwitchPubSubMessage struct {
	Type string `json:"type"`
	Data struct {
		Topic   string `json:"topic"`
		Message string `json:"message"`
	} `json:"data"`
}

type TwitchViewCountMessage struct {
	Type    string `json:"type"`
	Viewers int    `json:"viewers"`
}

func Dousetsu(ctx context.Context, loginID string) (*UserResponse, chan int, error) {
	gqlClient := graphql.NewClient(twitchGQLEndpoint)

	r := graphql.NewRequest(query)
	r.Var("login", loginID)
	r.Header.Add("Client-Id", twitchClientID)
	var resp UserResponse

	if err := gqlClient.Run(ctx, r, &resp); err != nil {
		return nil, nil, err
	}

	ch := make(chan int, 10)

	conn, _, err := websocket.DefaultDialer.Dial(twitchWSURL, nil)
	if err != nil {
		return nil, nil, err
	}

	log.Println("connected")

	listenMessage := map[string]interface{}{
		"type":  "LISTEN",
		"nonce": twitchPubSubNonce,
		"data": map[string]interface{}{
			"topics": []string{"video-playback-by-id." + resp.User.ID},
		},
	}
	if err := conn.WriteJSON(listenMessage); err != nil {
		return nil, nil, err
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pingMessage := map[string]string{"type": "PING"}
				if err := conn.WriteJSON(pingMessage); err != nil {
					log.Println("ping error:", err)
					return
				}
				log.Println("ping")
			case <-ctx.Done():
				log.Println("ping goroutine done")
				return
			}
		}
	}()

	go func() {
		defer conn.Close()
		defer close(ch)

		if resp.User.Stream.ID != "" {
			ch <- resp.User.Stream.ViewersCount
		}
		for {
			select {
			case <-ctx.Done():
				log.Println("message receive goroutine done")
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("read error:", err)
					return
				}

				var event TwitchPubSubMessage
				if err := json.Unmarshal(message, &event); err != nil {
					log.Println("unmarshal error:", err)
					continue
				}

				if event.Type == "MESSAGE" {
					var viewCount TwitchViewCountMessage
					if err := json.Unmarshal([]byte(event.Data.Message), &viewCount); err == nil {
						if viewCount.Type == "viewcount" {
							select {
							case ch <- viewCount.Viewers:
							case <-ctx.Done():
								return
							}
						}
					}
				}
			}
		}
	}()

	return &resp, ch, nil
}
