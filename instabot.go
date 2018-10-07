package main

func main() {
	// Gets the command line options
	parseOptions()
	// Gets the config
	getConfig()
	// Tries to login
	login()
	if *unfollow {
		syncFollowers()
	} else if *run {
		// Loop through tags ; follows, likes, and comments, according to the config file
		loopTags()
	}
	updateConfig()
}
