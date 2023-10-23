package flag

type GoubFlags struct {
	Help           bool `short:"h" long:"help" description:"Display usage" global:"true"`
	GenerateConfig struct {
		ClientID     int    `short:"i" long:"client_id" required:"true" description:"Client ID for the osu! API."`
		ClientSecret string `short:"s" long:"client_secret" required:"true" description:"Client secret for the osu! API."`
	} `command:"generate_config" description:"Generate a configuration file for osu! API."`
	Download struct {
		OutputDirectory string `short:"o" long:"output_directory" description:"Optional absolute path to the output folder. All maps will be saved there. Maps will be saved to the '/beatmaps/{target_user}' in the folder from which the program was called if not specified."`
		User            int    `short:"u" long:"user" required:"true" description:"Required! Numerical ID of the target user."`
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
}
