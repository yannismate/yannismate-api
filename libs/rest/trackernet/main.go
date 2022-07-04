package trackernet

type GetRankResponse struct {
	Rankings []Ranking `json:"rankings"`
}

type Ranking struct {
	Playlist Playlist `json:"playlist"`
	Mmr      int      `json:"mmr"`
	Rank     int      `json:"rank"`
	Division int      `json:"division"`
}

type Platform string

const (
	Steam = "steam"
	Epic  = "epic"
	PS    = "ps"
	Xbox  = "xbox"
)

type Playlist string

const (
	Unranked    = "unranked"
	Ranked1v1   = "ranked_1v1"
	Ranked2v2   = "ranked_2v2"
	Ranked3v3   = "ranked_3v3"
	Hoops       = "hoops"
	Rumble      = "rumble"
	Dropshot    = "dropshot"
	Snowday     = "snowday"
	Tournaments = "tournaments"
)
