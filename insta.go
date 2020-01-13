package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/spf13/viper"
)

// Storing user in session
var checkedUser = make(map[string]bool)

var insta *goinsta.Instagram

// login will try to reload a previous session, and will create a new one if it can't
func login() {
	err := reloadSession()
	if err != nil {
		createAndSaveSession()
	}
}

// reloadSession will attempt to recover a previous session
func reloadSession() error {

	insta, err := goinsta.Import("./goinsta-session")
	if err != nil {
		return errors.New("Couldn't recover the session")
	}

	if insta != nil {
		instabot.Insta = insta
	}

	log.Println("Successfully logged in")
	return nil

}

// Logins and saves the session
func createAndSaveSession() {
	insta := goinsta.New(viper.GetString("user.instagram.username"), viper.GetString("user.instagram.password"))
	instabot.Insta = insta
	err := instabot.Insta.Login()
	check(err)

	err = instabot.Insta.Export("./goinsta-session")
	check(err)
	log.Println("Created and saved the session")
}

func getInput(text string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(text)

	input, err := reader.ReadString('\n')
	check(err)
	return strings.TrimSpace(input)
}

// Checks if the user is in the slice
func contains(slice []goinsta.User, user goinsta.User) bool {
	for _, currentUser := range slice {
		if currentUser.Username == user.Username {
			return true
		}
	}
	return false
}

func (myInstabot MyInstabot) syncFollowers() {
	following := myInstabot.Insta.Account.Following()
	followers := myInstabot.Insta.Account.Followers()

	following.Next()
	followers.Next()

	var users []goinsta.User
	for _, user := range following.Users {
		if !contains(followers.Users, user) {
			users = append(users, user)
		}
	}
	fmt.Printf("\n%d users are not following you back!\n", len(users))
	answer := getInput("Do you want to unfollow these users? [yN]")
	if answer != "y" {
		fmt.Println("Not unfollowing.")
		os.Exit(0)
	}
	for _, user := range users {
		fmt.Printf("Unfollowing %s\n", user.Username)
		if !dev {
			user.Unfollow()
		}
		time.Sleep(6 * time.Second)
	}
}

// Follows a user, if not following already
func (myInstabot MyInstabot) followUser(user *goinsta.User) {
	log.Printf("Following %s\n", user.Username)
	err := user.FriendShip()
	check(err)
	// If not following already
	if !user.Friendship.Following {
		if !dev {
			user.Follow()
		}
		log.Println("Followed")
		numFollowed++
		report[line{tag, "follow"}]++
	} else {
		log.Println("Already following " + user.Username)
	}
}

func (myInstabot MyInstabot) loopTags() {
	for tag = range tagsList {
		limitsConf := viper.GetStringMap("tags." + tag)
		// Some converting
		limits = map[string]int{
			"follow":  int(limitsConf["follow"].(float64)),
			"like":    int(limitsConf["like"].(float64)),
			"comment": int(limitsConf["comment"].(float64)),
		}
		// What we did so far
		numFollowed = 0
		numLiked = 0
		numCommented = 0

		myInstabot.browse()
	}
	buildReport()
}

// Likes an image, if not liked already
func (myInstabot MyInstabot) likeImage(image goinsta.Item) {
	log.Println("Liking the picture")
	if !image.HasLiked {
		if !dev {
			image.Like()
		}
		log.Println("Liked")
		numLiked++
		report[line{tag, "like"}]++
	} else {
		log.Println("Image already liked")
	}
}

func (myInstabot MyInstabot) browse() {
	var i = 0
	for numFollowed < limits["follow"] || numLiked < limits["like"] || numCommented < limits["comment"] {
		log.Println("Fetching the list of images for #" + tag + "\n")
		i++

		// Getting all the pictures we can on the first page
		// Instagram will return a 500 sometimes, so we will retry 10 times.
		// Check retry() for more info.
		var images *goinsta.FeedTag
		err := retry(10, 20*time.Second, func() (err error) {
			images, err = myInstabot.Insta.Feed.Tags(tag)
			return
		})
		check(err)

		myInstabot.goThrough(images)

		if viper.IsSet("limits.maxRetry") && i > viper.GetInt("limits.maxRetry") {
			log.Println("Currently not enough images for this tag to achieve goals")
			break
		}
	}
}

// Goes through all the images for a certain tag
func (myInstabot MyInstabot) goThrough(images *goinsta.FeedTag) {
	var i = 1

	// do for other too
	for _, image := range images.Images {
		// Exiting the loop if there is nothing left to do
		if numFollowed >= limits["follow"] && numLiked >= limits["like"] && numCommented >= limits["comment"] {
			break
		}

		// Skip our own images
		if image.User.Username == viper.GetString("user.instagram.username") {
			continue
		}

		// Check if we should fetch new images for tag
		if i >= limits["follow"] && i >= limits["like"] && i >= limits["comment"] {
			break
		}

		// Skip checked user if the flag is turned on
		if checkedUser[image.User.Username] && noduplicate {
			continue
		}

		// Getting the user info
		// Instagram will return a 500 sometimes, so we will retry 10 times.
		// Check retry() for more info.

		var userInfo *goinsta.User
		err := retry(10, 20*time.Second, func() (err error) {
			userInfo, err = myInstabot.Insta.Profiles.ByName(image.User.Username)
			return
		})
		check(err)

		followerCount := userInfo.FollowerCount

		buildLine()

		checkedUser[userInfo.Username] = true
		log.Println("Checking followers for " + userInfo.Username + " - for #" + tag)
		log.Printf("%s has %d followers\n", userInfo.Username, followerCount)
		i++

		// Will only follow and comment if we like the picture
		like := followerCount > likeLowerLimit && followerCount < likeUpperLimit && numLiked < limits["like"]
		follow := followerCount > followLowerLimit && followerCount < followUpperLimit && numFollowed < limits["follow"] && like
		comment := followerCount > commentLowerLimit && followerCount < commentUpperLimit && numCommented < limits["comment"] && like

		// Checking if we are already following current user and skipping if we do
		skip := false
		following := myInstabot.Insta.Account.Following()

		following.Next()

		for _, user := range following.Users {
			if user.Username == userInfo.Username {
				skip = true
				break
			}
		}

		// Like, then comment/follow
		if !skip {
			if like {
				myInstabot.likeImage(image)
				if follow {
					myInstabot.followUser(userInfo)
				}
				if comment {
					//commentImage(image)
				}
			}
		}
		log.Printf("%s done\n\n", userInfo.Username)

		// This is to avoid the temporary ban by Instagram
		time.Sleep(20 * time.Second)
	}
}

// Comments an image (currently not working)
func commentImage(image goinsta.Item) {
	rand.Seed(time.Now().Unix())
	text := commentsList[rand.Intn(len(commentsList))]
	if !dev {
		image.Comments.Add(text)
	}
	log.Println("Commented " + text)
	numCommented++
	report[line{tag, "comment"}]++
}
