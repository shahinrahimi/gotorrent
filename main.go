package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gotorrent <infohash>")
		return
	}
	infohash := os.Args[1]

	downloadDir := filepath.Join("wd", "downloads")

	// Create the download directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		fmt.Println("Error creating download directory:", err)
		return
	}
	// Create a new torrent client
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = downloadDir
	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		fmt.Println("Error creating client", err)
	}
	defer client.Close()
	// Convert the infohash string to a metainfo.Hash
	hash := metainfo.NewHashFromHex(infohash)

	// Add the torrent using the infohash
	torrent, ok := client.AddTorrentInfoHash(hash)
	if !ok {
		fmt.Println("Error adding torrent:", err)
		return
	}
	<-torrent.GotInfo()
	fmt.Println("Torrent info loaded:", torrent.Name())

	torrent.DownloadAll()

	// handle gracefull shudown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Println("Shutting down...")
		client.Close()
		os.Exit(0)
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Print download progress
			progress := float64(torrent.BytesCompleted()) / float64(torrent.Info().TotalLength()) * 100
			fmt.Printf("Progress: %.2f%%\n", progress)

			// Check if download is complete
			if torrent.BytesCompleted() == torrent.Info().TotalLength() {
				fmt.Println("Download completed!")
				return
			}
		case <-torrent.Closed():
			fmt.Println("Torrent closed")
			return
		}
	}

}
