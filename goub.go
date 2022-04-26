package main

import (
	"fmt"
	"github.com/devfacet/gocmd/v3"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
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
	cfg.Section("OSU_SECRETS").NewKey("CLIENT_ID", strconv.FormatUint(uint64(clientID), 10))
	cfg.Section("OSU_SECRETS").NewKey("CLIENT_SECRET", clientSecret)

	_ = cfg.SaveTo(configDir)

	fmt.Printf("Successfully created the configuration file at: %s\n", configDir)
}

func CreateDefaultOutputFolders(username string) string {
	if _, err := os.Stat("beatmaps"); os.IsNotExist(err) {
		os.Mkdir("beatmaps", 0777)
	}

	outputDir := "beatmaps/" + username
	if _, err := os.Stat("beatmaps/" + username); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0777)
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

func DownloadMaps(beatmapsets map[uint]Beatmapset, outputDir string) {
	mapsDownloaded := 0
	for _, beatmapset := range beatmapsets {
		chimuURL := "https://api.chimu.moe/v1/download/" + strconv.FormatUint(uint64(beatmapset.ID), 10)
		resp, err := http.Get(chimuURL)
		if err != nil || resp.StatusCode != 200 {
			fmt.Printf("%d failed, please download manually.\n", beatmapset.ID)
			continue
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Reading %d body failed, please download manually.\n", beatmapset.ID)
		}

		r := regexp.MustCompile("[<>:\"/\\\\|?*]+")
		rawFileName := fmt.Sprintf("%d %s - %s.osz", beatmapset.ID, beatmapset.Artist, beatmapset.Title)
		fileName := r.ReplaceAllString(rawFileName, "")
		err = os.WriteFile(outputDir+"/"+fileName, data, 0777)
		if err != nil {
			fmt.Printf("%d failed, please download manually.\n", beatmapset.ID)
			log.Fatalln(err)
		}
		mapsDownloaded++
		fmt.Printf("Downloaded %d (%d/%d)\n", beatmapset.ID, mapsDownloaded, len(beatmapsets))
		time.Sleep(300 * time.Millisecond) // chimu is rate limited so i guess it has to be slower...

	}
	fmt.Printf("Download complete: Managed to download %d/%d maps.\n", mapsDownloaded, len(beatmapsets))
}

func main() {
	flags := struct {
		Help           bool `short:"h" long:"help" description:"Display usage" global:"true"`
		GenerateConfig struct {
			ClientID     uint   `short:"i" long:"client_id" required:"true" description:"Client ID for the osu! API."`
			ClientSecret string `short:"s" long:"client_secret" required:"true" description:"Client secret for the osu! API."`
		} `command:"generate_config" description:"Generate a configuration file for osu! API."`
		Download struct {
			OutputDirectory string `short:"o" long:"output_directory" description:"Optional absolute path to the output folder. All maps will be saved there. Maps will be saved to the '/beatmaps/{target_user}' in the folder from which the program was called if not specified."`
			User            uint   `short:"u" long:"user" required:"true" description:"Required! Numerical ID of the target user."`
			MostPlayed      bool   `short:"m" long:"most_played" description:"Download user's most played beatmaps."`
			Favorite        bool   `short:"f" long:"favorite" description:"Download user's favorite beatmaps."`
			Ranked          bool   `short:"r" long:"ranked" description:"Download user's ranked beatmaps."`
			Loved           bool   `short:"l" long:"loved" description:"Download user's loved beatmaps."`
			Pending         bool   `short:"p" long:"pending" description:"Download user's pending beatmaps."`
			Graveyard       bool   `short:"g" long:"graveyard" description:"Download user's graveyard beatmaps."`
		} `command:"download" description:"Download beatmaps from user's profile."`
	}{}

	gocmd.HandleFlag("GenerateConfig", func(cmd *gocmd.Cmd, args []string) error {
		clientId := flags.GenerateConfig.ClientID
		clientSecret := flags.GenerateConfig.ClientSecret
		CreateConfigurationFile(clientId, clientSecret)
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

		if !(mostPlayed || favorite || ranked || loved || pending || graveyard) {
			log.Fatalln("Please specify at least one beatmap type you want.")
		}

		clientID, clientSecret := GetSecretsFromConfig()
		client := CreateClient(clientID, clientSecret)

		user := client.GetUser(userID)

		if outputDir == "" {
			outputDir = CreateDefaultOutputFolders(user.Username)
		} else if err := ErrorIfOutputDirDoesNotExist(outputDir); err != nil {
			log.Fatalf("Output folder does not exist:\n%s", err)
		}

		beatmapCountMap := map[string]uint{
			MOST_PLAYED: user.BeatmapPlaycountsCount,
			FAVOURITE:   user.FavoriteBeatmapsetCount,
			RANKED:      user.RankedBeatmapsetCount,
			LOVED:       user.LovedBeatmapsetCount,
			PENDING:     user.PendingBeatmapsetCount,
			GRAVEYARD:   user.GraveyardBeatmapsetCount,
		}
		beatmapTypesToGet := map[string]bool{
			MOST_PLAYED: mostPlayed,
			FAVOURITE:   favorite,
			RANKED:      ranked,
			LOVED:       loved,
			PENDING:     pending,
			GRAVEYARD:   graveyard,
		}
		beatmapsets := client.GetBeatmapsetsForUser(user.ID, beatmapTypesToGet, beatmapCountMap)
		DownloadMaps(beatmapsets, outputDir)
		return nil
	})

	gocmd.New(gocmd.Options{
		Name:        "gopbd",
		Description: "osu! Profile Beatmap Downloader",
		Flags:       &flags,
		ConfigType:  gocmd.ConfigTypeAuto,
	})
}
