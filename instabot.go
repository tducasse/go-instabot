package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/ahmdrz/goinsta"
	"github.com/ahmdrz/goinsta/response"
	"github.com/spf13/viper"
)

func main() {
	// Comment the next section if you don't want to log events in a file
	// ------------------------------ SECTION ------------------------------
	// Opens a log file
	t := time.Now()
	logFile, err := os.OpenFile("instabot-"+t.Format("2006-01-02-15-04-05")+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	// Duplicates the writer to stdout and logFile
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// -------------------------------- END --------------------------------

	// This is the config file
	viper.SetConfigFile("./config/config.json")

	// Reads the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Confirm which config file is used
	log.Printf("Using config: %s\n\n", viper.ConfigFileUsed())

	// Creates a new instance of the API for instagram
	insta := goinsta.New(viper.GetString("user.instagram.username"), viper.GetString("user.instagram.password"))

	// An image will be liked if the poster has more followers than likeLowerLimit, and less than likeUpperLimit
	likeLowerLimit := viper.GetInt("limits.like.min")
	likeUpperLimit := viper.GetInt("limits.like.max")

	// A user will be followed if he has more followers than followLowerLimit, and less than followUpperLimit
	// Needs to be a subset of the like interval
	followLowerLimit := viper.GetInt("limits.follow.min")
	followUpperLimit := viper.GetInt("limits.follow.max")

	// An image will be commented if the poster has more followers than commentLowerLimit, and less than commentUpperLimit
	// Needs to be a subset of the like interval
	commentLowerLimit := viper.GetInt("limits.comment.min")
	commentUpperLimit := viper.GetInt("limits.comment.max")

	// Hashtags list. Do not put the '#' in the config file
	tagsList := viper.GetStringMap("tags")

	// Comments list
	commentsList := viper.GetStringSlice("comments")

	// Report is a struct to store the report
	type Report struct {
		Tag, Action string
	}

	// Report that will be sent at the end of the script
	report := make(map[Report]int)

	// Tries to login to Instagram
	if err := insta.Login(); err != nil {
		panic(err)
	}

	defer insta.Logout()

	// Go through all the tags in the list
	for tag := range tagsList {
		limitsConf := viper.GetStringMap("tags." + tag)
		// Some converting
		limits := map[string]int{
			"follow":  int(limitsConf["follow"].(float64)),
			"like":    int(limitsConf["like"].(float64)),
			"comment": int(limitsConf["comment"].(float64)),
		}

		// What we did so far
		numFollowed := 0
		numLiked := 0
		numCommented := 0

		// While we haven't met the requirements
		for numFollowed < limits["follow"] || numLiked < limits["like"] || numCommented < limits["comment"] {
			log.Println("Fetching the list of images for #" + tag + "\n")

			// Getting all the pictures we can on the first page
			// Instagram will return a 500 sometimes, so we will retry 10 times.
			// Check retry() for more info.
			var images response.TagFeedsResponse
			err := retry(10, 20*time.Second, func() (err error) {
				images, err = insta.TagFeed(tag)
				return
			})
			checkErr(err)

			// Go through all the images in the response
			for _, image := range images.FeedsResponse.Items {
				// Exiting the loop if there is nothing left to do
				if numFollowed >= limits["follow"] && numLiked >= limits["like"] && numCommented >= limits["comment"] {
					break
				}

				// Getting the user info
				// Instagram will return a 500 sometimes, so we will retry 10 times.
				// Check retry() for more info.
				var posterInfo response.GetUsernameResponse
				err := retry(10, 20*time.Second, func() (err error) {
					posterInfo, err = insta.GetUserByID(image.User.ID)
					return
				})
				checkErr(err)

				poster := posterInfo.User
				followerCount := poster.FollowerCount

				// Builds the line for the report
				reportTag := ""
				for index, element := range report {
					if index.Tag == tag {
						reportTag += fmt.Sprintf("%s %d/%d - ", index.Action, element, limits[index.Action])
					}
				}

				// Prints the report line on the screen / in the log file
				if reportTag != "" {
					log.Println(strings.TrimSuffix(reportTag, " - "))
				}

				log.Println("Checking followers for " + poster.Username + " - for #" + tag)
				log.Printf("%s has %d followers\n", poster.Username, followerCount)

				// Will only follow and comment if we like the picture
				like := followerCount > likeLowerLimit && followerCount < likeUpperLimit && numLiked < limits["like"]
				follow := followerCount > followLowerLimit && followerCount < followUpperLimit && numFollowed < limits["follow"] && like
				comment := followerCount > commentLowerLimit && followerCount < commentUpperLimit && numCommented < limits["comment"] && like

				// Like, then comment/follow
				if like {
					log.Println("Liking the picture")
					if !image.HasLiked {
						insta.Like(image.ID)
						log.Println("Liked")
						numLiked++
						report[Report{tag, "like"}]++
						// Follow
						if follow {
							log.Printf("Following %s\n", poster.Username)
							userFriendShip, err := insta.UserFriendShip(poster.ID)
							checkErr(err)
							// If not following already
							if !userFriendShip.Following {
								insta.Follow(poster.ID)
								log.Println("Followed")
								numFollowed++
								report[Report{tag, "follow"}]++
							} else {
								log.Println("Already following " + poster.Username)
							}
						}
						// Comment
						if comment {
							rand.Seed(time.Now().Unix())
							text := commentsList[rand.Intn(len(commentsList))]
							insta.Comment(image.ID, text)
							log.Println("Commented " + text)
							numCommented++
							report[Report{tag, "comment"}]++
						}
					} else {
						log.Println("Image already liked")
					}
				}
				log.Printf("%s done\n\n", poster.Username)

				// This is to avoid the temporary ban by Instagram
				time.Sleep(20 * time.Second)
			}
		}
	}

	// Builds the report
	reportAsString := ""
	for index, element := range report {
		var times string
		if element == 1 {
			times = "time"
		} else {
			times = "times"
		}
		if index.Action == "like" {
			reportAsString += fmt.Sprintf("#%s has been liked %d %s\n", index.Tag, element, times)
		} else {
			reportAsString += fmt.Sprintf("#%s has been %sed %d %s\n", index.Tag, index.Action, element, times)
		}
	}

	// Displays the report on the screen / log file
	fmt.Println(reportAsString)

	// Sends the report to the email in the config file
	send(reportAsString, true)
}

// Send an email. Check out the "mail" section of the "config.json" file.
func send(body string, success bool) {
	from := viper.GetString("user.mail.from")
	pass := viper.GetString("user.mail.password")
	to := viper.GetString("user.mail.to")

	status := func() string {
		if success {
			return "Success!"
		}
		return "Failure!"
	}()
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject:" + status + "  go-instabot\n\n" +
		body

	err := smtp.SendMail(viper.GetString("user.mail.smtp"),
		smtp.PlainAuth("", from, pass, viper.GetString("user.mail.server")),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("sent")
}

// Retry the same function [function], a certain number of times (maxAttempts).
// It is exponential : the 1st time it will be (sleep), the 2nd time, (sleep) x 2, the 3rd time, (sleep) x 3, etc.
// If this function fails to recover after an error, it will send an email to the address in the config file.
func retry(maxAttempts int, sleep time.Duration, function func() error) (err error) {
	for currentAttempt := 0; currentAttempt < maxAttempts; currentAttempt++ {
		err = function()
		if err == nil {
			return
		}
		for i := 0; i <= currentAttempt; i++ {
			time.Sleep(sleep)
		}
		log.Println("Retrying after error:", err)
	}

	send(fmt.Sprintf("The script has stopped due to an unrecoverable error :\n%s", err), false)

	return fmt.Errorf("After %d attempts, last error: %s", maxAttempts, err)
}

// CheckErr checks if there is an error, and log.Fatal() if need be.
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
