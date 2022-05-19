package main

import (
	"fmt"
	"github.com/devfacet/gocmd/v3"
	"gopkg.in/ini.v1"
	"goub/api"
	"log"
	"os"
)

func CreateConfigurationFile(clientID uint, clientSecret string) {
	fmt.Println("Creating configuration file in your home directory.")
	var configDir string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to get user home directory:\n%s", err)
	}

	configDir = homeDir + "/goub-config.ini"
	file, _ := os.Create(configDir) // Creates or truncates the file
	file.Close()

	cfg, err := ini.Load(configDir)
	if err != nil {
		log.Fatalf("Unable to read configuration file:\n%s", err)
	}

	cfg.NewSection("OSU_SECRETS")
	cfg.Section("OSU_SECRETS").NewKey("CLIENT_ID", fmt.Sprintf("%d", clientID))
	cfg.Section("OSU_SECRETS").NewKey("CLIENT_SECRET", clientSecret)

	_ = cfg.SaveTo(configDir)

	fmt.Printf("Successfully created the configuration file at: %s\n", configDir)
}

func CreateDefaultOutputFolders(username string) string {
	if _, err := os.Stat("beatmaps"); os.IsNotExist(err) {
		os.Mkdir("beatmaps", 0666)
	}

	outputDir := "beatmaps/" + username
	if _, err := os.Stat("beatmaps/" + username); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0666)
	}

	return outputDir
}

func ErrorIfOutputDirDoesNotExist(outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return err
	}
	return nil
}

func GetSecretsFromConfig() (uint, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to get user home directory:\n%s", err)
	}

	cfg, err := ini.Load(homeDir + "/goub-config.ini")
	if err != nil {
		log.Fatalf("Unable to read configuration file:\n%s", err)
	}

	clientID, err := cfg.Section("OSU_SECRETS").Key("CLIENT_ID").Uint()
	if err != nil {
		log.Fatalf("Unable to parse client ID to uint from configuration file:\n%s", err)
	}
	clientSecret := cfg.Section("OSU_SECRETS").Key("CLIENT_SECRET").String()

	return clientID, clientSecret
}

func CreateFolders(user api.User, outputDir string) string {
	if outputDir == "" {
		return CreateDefaultOutputFolders(user.Username)
	} else if err := ErrorIfOutputDirDoesNotExist(outputDir); err != nil {
		log.Fatalf("Output folder does not exist:\n%s", err)
	}

	return outputDir
}

func GetBeatmapCountsForUser(user api.User) map[string]uint {
	return map[string]uint{
		api.MostPlayed: user.BeatmapPlaycountsCount,
		api.Favorite:   user.FavoriteBeatmapsetCount,
		api.Ranked:     user.RankedBeatmapsetCount,
		api.Loved:      user.LovedBeatmapsetCount,
		api.Pending:    user.PendingBeatmapsetCount,
		api.Graveyard:  user.GraveyardBeatmapsetCount,
		api.Firsts:     user.ScoresFirstCount,
		api.Best:       user.ScoresBestCount,
	}
}

func main() {
	flags := struct {
		Help           bool `short:"h" long:"help" description:"Display usage" global:"true"`
		GenerateConfig struct {
			ClientID     uint   `short:"i" long:"client_id" required:"true" description:"Client ID for the osu! API."`
			ClientSecret string `short:"s" long:"client_secret" required:"true" description:"Client secret for the osu! API."`
		} `command:"generate_config" description:"Generate a configuration file for osu! API."`
		RecursiveFavorites struct {
			OutputDirectory string `short:"o" long:"output_directory" description:"Optional absolute path to the output folder. All maps will be saved there. Maps will be saved to the '/beatmaps/{target_user}' in the folder from which the program was called if not specified."`
			RecursionDepth  uint   `short:"d" long:"depth" description:"Optional recursion depth. While it is optional, it is recommended to specify one or else it will run for a very long time. Recommended value is 3."`
			StartUser       uint   `short:"u" long:"user" required:"true" description:"Required! Numerical ID of the start user."`
		} `command:"recursive_favorites" description:"Download user's favorites and his favorite maps' authors' favorites and etc until there is none left."`
		Download struct {
			OutputDirectory string `short:"o" long:"output_directory" description:"Optional absolute path to the output folder. All maps will be saved there. Maps will be saved to the '/beatmaps/{target_user}' in the folder from which the program was called if not specified."`
			User            uint   `short:"u" long:"user" required:"true" description:"Required! Numerical ID of the target user."`
			MostPlayed      bool   `short:"m" long:"most_played" description:"Download user's most played beatmaps."`
			Favorite        bool   `short:"f" long:"favorite" description:"Download user's favorite beatmaps."`
			Ranked          bool   `short:"r" long:"ranked" description:"Download user's ranked beatmaps."`
			Loved           bool   `short:"l" long:"loved" description:"Download user's loved beatmaps."`
			Pending         bool   `short:"p" long:"pending" description:"Download user's pending beatmaps."`
			Graveyard       bool   `short:"g" long:"graveyard" description:"Download user's graveyard beatmaps."`
			Best            bool   `short:"b" long:"best" description:"Download user's top play beatmaps."`
			Firsts          bool   `short:"1" long:"firsts" description:"Download user's beatmaps where they hold the first place."`
			GameMode        string `long:"gamemode" description:"Specify game mode if downloading best or firsts. Choose: fruits, mania, osu, taiko." default:"osu" args:"osu,taiko,mania,fruits" allow-unknown-arg:"false"`
		} `command:"download" description:"Download beatmaps from user's profile."`
	}{}

	gocmd.HandleFlag("GenerateConfig", func(cmd *gocmd.Cmd, args []string) error {
		clientId := flags.GenerateConfig.ClientID
		clientSecret := flags.GenerateConfig.ClientSecret
		CreateConfigurationFile(clientId, clientSecret)
		return nil
	})

	gocmd.HandleFlag("RecursiveFavorites", func(cmd *gocmd.Cmd, args []string) error {
		startUserID := flags.RecursiveFavorites.StartUser
		outputDir := flags.RecursiveFavorites.OutputDirectory
		recursionDepthLimit := flags.RecursiveFavorites.RecursionDepth

		clientID, clientSecret := GetSecretsFromConfig()
		client := api.CreateClient(clientID, clientSecret)
		user := client.GetUser(startUserID)
		outputDir = CreateFolders(user, outputDir)

		beatmapsetsForUsers := client.GetBeatmapIDsForRecursiveFavorites(user, make(map[api.User]map[uint]api.Beatmapset), 0, recursionDepthLimit)
		beatmapsetsToDownload := make(map[uint]api.Beatmapset)
		for _, userBeatmapsets := range beatmapsetsForUsers {
			for _, beatmapset := range userBeatmapsets {
				beatmapsetsToDownload[beatmapset.ID] = beatmapset
			}
		}
		client.DownloadMaps(beatmapsetsToDownload, outputDir)
		return nil
	})

	gocmd.HandleFlag("Download", func(cmd *gocmd.Cmd, args []string) error {
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

		clientID, clientSecret := GetSecretsFromConfig()
		client := api.CreateClient(clientID, clientSecret)

		user := client.GetUser(userID)

		outputDir = CreateFolders(user, outputDir)

		beatmapCountMap := GetBeatmapCountsForUser(user)
		beatmapTypesToGet := map[string]bool{
			api.MostPlayed: mostPlayed,
			api.Favorite:   favorite,
			api.Ranked:     ranked,
			api.Loved:      loved,
			api.Pending:    pending,
			api.Graveyard:  graveyard,
			api.Best:       best,
			api.Firsts:     firsts,
		}
		beatmapsets := client.GetBeatmapsetsForUser(user.ID, beatmapTypesToGet, beatmapCountMap, gameMode)
		client.DownloadMaps(beatmapsets, outputDir)
		return nil
	})

	gocmd.New(gocmd.Options{
		Name:        "gopbd",
		Description: "osu! Profile Beatmap Downloader",
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
	})
}
