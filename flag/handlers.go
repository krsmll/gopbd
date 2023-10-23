package flag

import (
	"fmt"
	"log"

	"github.com/devfacet/gocmd/v3"
	"goub/config"
	"goub/osu"
)

func HandleCreateConfig(flags *GoubFlags) func(cmd *gocmd.Cmd, args []string) error {
	return func(cmd *gocmd.Cmd, args []string) error {
		clientId := flags.GenerateConfig.ClientID
		clientSecret := flags.GenerateConfig.ClientSecret
		config.Create(clientId, clientSecret)
		return nil
	}
}

func HandleDownload(flags *GoubFlags) func(cmd *gocmd.Cmd, args []string) error {
	return func(cmd *gocmd.Cmd, args []string) error {
		outputDir := flags.Download.OutputDirectory
		userID := flags.Download.User
		mostPlayed := flags.Download.MostPlayed
		favorite := flags.Download.Favorite
		ranked := flags.Download.Ranked
		loved := flags.Download.Loved
		pending := flags.Download.Pending
		graveyard := flags.Download.Graveyard
		best := flags.Download.Best
		firsts := flags.Download.Firsts
		gameMode := flags.Download.GameMode

		if !(mostPlayed || favorite || ranked || loved || pending || graveyard || best || firsts) {
			log.Fatalln("Please specify at least one beatmap type you want.")
		}

		clientID, clientSecret := config.GetSecrets()
		client := osu.NewClient(clientID, clientSecret)

		user := client.GetUser(userID)

		outputDir = config.CreateFolders(user, outputDir)

		beatmapCountMap := user.GetBeatmapCounts()
		beatmapTypesToGet := map[osu.BeatmapType]bool{
			osu.BeatmapTypeMostPlayed: mostPlayed,
			osu.BeatmapTypeFavorite:   favorite,
			osu.BeatmapTypeRanked:     ranked,
			osu.BeatmapTypeLoved:      loved,
			osu.BeatmapTypePending:    pending,
			osu.BeatmapTypeGraveyard:  graveyard,
			osu.BeatmapTypeBest:       best,
			osu.BeatmapTypeFirsts:     firsts,
		}

		var total int
		for t, v := range beatmapTypesToGet {
			if v {
				total += beatmapCountMap[t]
			}
		}

		fmt.Println(user.Username)
		ch := make(chan int, total)
		go client.StartGatheringBeatmapsets(ch, userID, beatmapTypesToGet, beatmapCountMap, osu.GamemodeFromString(gameMode))
		client.Download(ch, outputDir, total)
		return nil
	}
}
