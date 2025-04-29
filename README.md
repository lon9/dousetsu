# Dousetsu Twitch Viewer Count Monitor

This project is a Twitch viewer count monitor built using the Fyne GUI toolkit. It displays the current viewer count, user information, and a graph of viewer count history.

## Features

- Display current viewer count
- Show user information including display name, profile image, followers count, stream title, and game name
- Indicate viewer count changes with arrows
- Display the last update time
- Plot a graph of viewer count history

## Requirements

- Go 1.12 or later
- Fyne v2.0 or later

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/lon9/dousetsu.git
   cd dousetsu
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

## Usage

1. Build the application:

   ```sh
   go build -o monitor ./monitor/main.go
   ```

2. Run the application with your Twitch login ID:

   ```sh
   ./monitor <Twitch login ID>
   ```

## Example

To run the application for the Twitch user `xqc`, use the following command:

```sh
./monitor xqc
```
