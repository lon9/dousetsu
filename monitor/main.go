package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/lon9/dousetsu"
)

func main() {
	// Get command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: monitor <Twitch login ID>")
		os.Exit(1)
	}
	loginID := os.Args[1]

	a := app.New()
	w := a.NewWindow("Twitch Viewer Count Monitor")

	viewerCount := binding.NewString()
	viewerCount.Set("Viewer Count: 0")
	viewerCountText := canvas.NewText("Viewer Count: 0", nil)
	viewerCountText.TextSize = 36 // Set font size larger

	arrowText := canvas.NewText("", nil)
	arrowText.TextSize = 36 // Set arrow font size larger

	updateTimeLabel := widget.NewLabel("Last Update: -")

	userNameLabel := widget.NewLabel("User: ")
	userImage := canvas.NewImageFromResource(nil)
	userImage.FillMode = canvas.ImageFillContain
	userImage.SetMinSize(fyne.NewSize(300, 300)) // Set image size

	followersLabel := widget.NewLabel("Followers: 0")
	streamTitleLabel := widget.NewLabel("Stream Title: ")
	gameNameLabel := widget.NewLabel("Game: ")

	// Create container for the graph
	graphContainer := container.NewWithoutLayout()

	w.SetContent(container.NewVBox(
		viewerCountText,
		arrowText,
		updateTimeLabel,
		userNameLabel,
		userImage,
		followersLabel,
		streamTitleLabel,
		gameNameLabel,
		graphContainer,
	))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userResponse, viewerCountChan, err := dousetsu.Dousetsu(ctx, loginID)
	if err != nil {
		log.Fatalf("Failed to start Dousetsu: %v", err)
	}

	log.Printf("User: %s", userResponse.User.Login)

	// Set user name and profile image
	userNameLabel.SetText(fmt.Sprintf("User: %s", userResponse.User.DisplayName))
	followersLabel.SetText(fmt.Sprintf("Followers: %d", userResponse.User.Followers.TotalCount))
	if userResponse.User.Stream.ID != "" {
		streamTitleLabel.SetText(fmt.Sprintf("Stream Title: %s", userResponse.User.Stream.Title))
		gameNameLabel.SetText(fmt.Sprintf("Game: %s", userResponse.User.Stream.Game.Name))
	}

	go func() {
		uri, err := storage.ParseURI(userResponse.User.ProfileImageURL)
		if err != nil {
			log.Println("Failed to parse URI:", err)
			return
		}

		userImage.Resource = canvas.NewImageFromURI(uri).Resource
		userImage.Refresh()
	}()

	var previousCount int
	var viewerCounts []int

	go func() {
		for count := range viewerCountChan {
			log.Println("Viewer Count:", count)
			viewerCount.Set(fmt.Sprintf("Viewer Count: %d", count))
			viewerCountText.Text = fmt.Sprintf("Viewer Count: %d", count)

			if previousCount < count {
				arrowText.Text = "↑"
				arrowText.Color = color.RGBA{0, 255, 0, 255} // Green color
			} else if previousCount > count {
				arrowText.Text = "↓"
				arrowText.Color = color.RGBA{255, 0, 0, 255} // Red color
			} else {
				arrowText.Text = ""
			}

			previousCount = count

			viewerCountText.Refresh()
			arrowText.Refresh()

			// Set update time
			updateTimeLabel.SetText(fmt.Sprintf("Last Update: %s", time.Now().Format("2006-01-02 15:04:05")))

			// Save viewer count history
			viewerCounts = append(viewerCounts, count)
			if len(viewerCounts) > 100 {
				viewerCounts = viewerCounts[1:]
			}

			// Adjust graph scaling
			maxCount := 1
			for _, v := range viewerCounts {
				if v > maxCount {
					maxCount = v
				}
			}

			// Update graph
			graphContainer.RemoveAll()
			for i := 1; i < len(viewerCounts); i++ {
				line := canvas.NewLine(color.RGBA{0, 0, 255, 255})
				line.StrokeWidth = 2
				line.Position1 = fyne.NewPos(float32(i-1)*4, 100-float32(viewerCounts[i-1])*100/float32(maxCount))
				line.Position2 = fyne.NewPos(float32(i)*4, 100-float32(viewerCounts[i])*100/float32(maxCount))
				graphContainer.Add(line)
			}
			graphContainer.Refresh()
		}
	}()

	w.Resize(fyne.NewSize(400, 800))
	w.ShowAndRun()
}
