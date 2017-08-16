package main

func main() {
	// Gets the command line options
	parseOptions()
	// Gets the config
	getConfig()
	// Tries to login
	login()
	// Loop through tags ; follows, likes, and comments, according to the config file
	loopTags()
}
