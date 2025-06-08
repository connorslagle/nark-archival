package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

func main() {
	relay := khatru.NewRelay()

	relay.Info.Name = "NARK Archival Relay"
	relay.Info.Description = "A NOSTR archival relay for academic content"
	relay.Info.PubKey = ""
	relay.Info.Contact = ""

	relay.QueryEvents = append(relay.QueryEvents, func(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
		ch := make(chan *nostr.Event)
		
		go func() {
			defer close(ch)
		}()
		
		return ch, nil
	})

	relay.StoreEvent = append(relay.StoreEvent, func(ctx context.Context, event *nostr.Event) error {
		log.Printf("Storing event: %s", event.ID)
		return nil
	})

	relay.DeleteEvent = append(relay.DeleteEvent, func(ctx context.Context, event *nostr.Event) error {
		return fmt.Errorf("deletion not allowed on archival relay")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3334"
	}

	fmt.Printf("Running NARK archival relay on :%s\n", port)
	if err := relay.Start("0.0.0.0", port); err != nil {
		log.Fatal(err)
	}
}