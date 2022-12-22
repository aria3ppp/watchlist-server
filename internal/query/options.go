package query

type Options struct {
	Offset    int
	Limit     int
	SortField string
	SortOrder string
}

type SortOrderOptions struct {
	Offset    int
	Limit     int
	SortOrder string
}

type SearchOptions struct {
	Query string
	From  int
	Size  int
}

type WatchlistOptions struct {
	Offset           int
	Limit            int
	SortOrder        string
	WhereTimeWatched string
}
