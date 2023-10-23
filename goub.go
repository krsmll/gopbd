package main

import (
	"github.com/devfacet/gocmd/v3"
	"goub/flag"
)

func main() {
	flags := flag.GoubFlags{}

	gocmd.HandleFlag("GenerateConfig", flag.HandleCreateConfig(&flags))
	gocmd.HandleFlag("Download", flag.HandleDownload(&flags))

	gocmd.New(gocmd.Options{
		Name:        "goub",
		Description: "osu! User Beatmap Downloader",
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
	})
}
