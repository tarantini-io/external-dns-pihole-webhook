package main

import (
	"fmt"
	"github.com/tarantini-io/external-dns-pihole-webhook/cmd/webhook/configuration"
	"github.com/tarantini-io/external-dns-pihole-webhook/cmd/webhook/dnsprovider"
	"github.com/tarantini-io/external-dns-pihole-webhook/cmd/webhook/log"
	"github.com/tarantini-io/external-dns-pihole-webhook/cmd/webhook/server"

	"github.com/tarantini-io/external-dns-pihole-webhook/pkg/webhook"
	"moul.io/banner"

	"go.uber.org/zap"
)

var (
	Version = "local"
	Gitsha  = "?"
)

func main() {
	fmt.Println(banner.Inline("pihole"))
	fmt.Println(banner.Inline(fmt.Sprintf("version: %s gitsha: %s", Version, Gitsha)))

	log.Init()

	config := configuration.Init()
	provider, err := dnsprovider.Init(config)
	if err != nil {
		log.Fatal("failed to initialize provider", zap.Error(err))
	}

	main, health := server.Init(config, webhook.New(provider))
	server.ShutdownGracefully(main, health)
}
