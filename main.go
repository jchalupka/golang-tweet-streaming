package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type twitterKeys struct {
	consumerKey    *string
	consumerSecret *string
	accessToken    *string
	accessSecret   *string
}

func getTwitterFlags() twitterKeys {
	flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
	consumerKey := flags.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flags.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flags.String("access-token", "", "Twitter Access Token")
	accessSecret := flags.String("access-secret", "", "Twitter Access Secret")
	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "TWITTER")

	if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
		log.Fatal("Consumer key/secret and Access token/secret required")
	}

	return twitterKeys{consumerKey, consumerSecret, accessToken, accessSecret}
}

func GetTweetStreamer() *twitter.Stream {
	twitterFlags := getTwitterFlags()

	config := oauth1.NewConfig(*twitterFlags.consumerKey, *twitterFlags.consumerSecret)
	token := oauth1.NewToken(*twitterFlags.accessToken, *twitterFlags.accessSecret)

	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter Client
	client := twitter.NewClient(httpClient)

	// Filter
	filterParams := &twitter.StreamFilterParams{
		StallWarnings: twitter.Bool(true),
		Locations:     []string{"-180,-90,180,90"}, // All locations
	}

	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	return stream
}

func main() {
	stream := GetTweetStreamer()
	demux := twitter.NewSwitchDemux()

	demux.Tweet = func(tweet *twitter.Tweet) {
		fmt.Printf("%+v\n\n", tweet)
	}

	// Start printing
	go demux.HandleChan(stream.Messages)

	// Stop on ctrl-c
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	fmt.Println("Stopping Stream...")
	stream.Stop()
}
