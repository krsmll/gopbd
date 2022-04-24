package gosu

type BeatmapType string

const (
	FAVOURITE   BeatmapType = "favourite"
	GRAVEYARD               = "graveyard"
	LOVED                   = "loved"
	MOST_PLAYED             = "most_played"
	RANKED                  = "ranked"
	PENDING                 = "pending"
)

func (b BeatmapType) String() string {
	switch b {
	case FAVOURITE:
		return "favourite"
	case GRAVEYARD:
		return "graveyard"
	case LOVED:
		return "loved"
	case MOST_PLAYED:
		return "most_played"
	case PENDING:
		return "pending"
	default:
		return "ranked"
	}
}
