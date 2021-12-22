package twitterspy

type TwitterUserInfo struct {
	ID     string `json:"id"`
	RestID string `json:"rest_id"`
	Legacy struct {
		CreatedAt            string   `json:"created_at"`
		Description          string   `json:"description"`
		FastFollowersCount   int      `json:"fast_followers_count"`
		FavouritesCount      int      `json:"favourites_count"`
		FollowersCount       int      `json:"followers_count"`
		FriendsCount         int      `json:"friends_count"`
		ListedCount          int      `json:"listed_count"`
		StatusesCount        int      `json:"statuses_count"`
		Location             string   `json:"location"`
		MediaCount           int      `json:"media_count"`
		Name                 string   `json:"name"`
		NormalFollowersCount int      `json:"normal_followers_count"`
		PinnedTweetIdsStr    []string `json:"pinned_tweet_ids_str"`
		ProfileBannerURL     string   `json:"profile_banner_url"`
		ProfileImageURLHTTPS string   `json:"profile_image_url_https"`
		ScreenName           string   `json:"screen_name"`
		URL                  string   `json:"url"`
		Verified             bool     `json:"verified"`
	} `json:"legacy"`
}
