package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/ini.v1"
	"goub/osu"
)

func Create(clientID int, clientSecret string) {
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
		os.Mkdir("beatmaps", 0o777)
	}

	outputDir := "beatmaps/" + username
	if _, err := os.Stat("beatmaps/" + username); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0o777)
	}

	return outputDir
}

func ErrorIfOutputDirDoesNotExist(outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return err
	}
	return nil
}

func GetSecrets() (int, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to get user home directory:\n%s", err)
	}

	cfg, err := ini.Load(homeDir + "/goub-config.ini")
	if err != nil {
		log.Fatalf("Unable to read configuration file:\n%s", err)
	}

	clientID, err := cfg.Section("OSU_SECRETS").Key("CLIENT_ID").Int()
	if err != nil {
		log.Fatalf("Unable to parse client ID to int from configuration file:\n%s", err)
	}
	clientSecret := cfg.Section("OSU_SECRETS").Key("CLIENT_SECRET").String()

	return clientID, clientSecret
}

func CreateFolders(user osu.User, outputDir string) string {
	if outputDir == "" {
		return CreateDefaultOutputFolders(user.Username)
	} else if err := ErrorIfOutputDirDoesNotExist(outputDir); err != nil {
		log.Fatalf("Output folder does not exist:\n%s", err)
	}

	return outputDir
}
