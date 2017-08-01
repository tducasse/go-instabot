[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0) [![Made with: Golang](https://img.shields.io/badge/Made%20with-Golang-brightgreen.svg)](https://golang.org/)
# What is go-instabot?

The easiest way to boost your Instagram account and get likes and followers.

Go-instabot automates following users, liking pictures, and commenting on Instagram. It uses the unofficial but excellent Go Instagram API, [goinsta](https://github.com/ahmdrz/goinsta).

![Instabot demo gif](/docs/instabot.gif)

### Concept
The idea behind the script is that when you like, follow, or comment something, it will draw the user's attention back to your own account. There's a hidden convention in Instagram, that will make people follow you back, as a way of saying "thank you" I guess.


# How does it work?

- There is a config file, where you can enter hashtags, and for each hashtag, how many likes, comments, and follows you want.
- The script will then fetch pictures from the 'explore by tags' page on Instagram, using the API.
- It will decide (based on your settings) if it has to follow, like or comment.
- At the end of the routine, an email will be sent, with a report on everything that's been done.

The script is coded so that your Instagram account will not get banned ; it waits between every call to simulate human behavior.

Additionally, there is a retry mechanism in the eventuality that Instagram is too slow to answer (it can happen sometimes), and the script will wait for some time before trying again.

# How to use
## Installation

1. [Install Go](https://golang.org/doc/install) on your system.

2. Download go-instabot, by executing this command in your terminal / cmd :

   `go get github.com/tducasse/go-instabot`

## Configuration
### Config.json
Go to the project folder :

`cd [YOUR_GO_PATH]/src/github.com/tducasse/go-instabot`

There, in the 'dist/' folder, you will find a sample 'config.json', that you have to copy to the 'config/' folder :

```go
{
    "user" : {
        "instagram" : {
            "username" : "foobar",          // your instagram username
            "password" : "fooBAR"           // your instagram password
        },
        "mail" : {
            "from" : "foo@bar.foo",         // the address from which you want to send the emails
            "password" : "foOb@r",          // the password for this email address
            "to" : "bar@foo.bar",           // the address to which you want to send the emails
            "smtp" : "smtp.gmail.com:587",  // the smtp address for your mail server
            "server" : "smtp.gmail.com"     // the domain name for your server
        }
    },
    "limits" : {                            // this sets when the script will choose to do something
        "like" : {                          // the script will like a picture only if :
            "min" : 0,                      //      - the user has more than 0 followers
            "max" : 10000                   //      - the user has less than 10000 followers  
        },
        "comment" : {                       // the script will comment only if :
            "min" : 100,                    //      - the user has more than 100 followers
            "max" : 10000                   //      - the user has less than 10000 followers  
        },
        "follow" : {                        // the script will follow someone only if :
            "min" : 200,                    //      - the user has more than 200 followers
            "max" : 10000                   //      - the user has less than 10000 followers  
        }
    },
    "tags" : {                              // this is the list of hashtags you want to explore
        "dog" : {                           // do not put the '#' symbol
            "like" : 3,                     // the number you want to like
            "comment" : 2,                  // the number you want to comment
            "follow" : 1                    // the number you want to follow
        },
        "cat" : {                           // another hashtag ('#cat')
            "like" : 3,
            "comment" : 2,
            "follow" : 1
        }                                   // following these examples, add as many as you want
    },
    "comments" : [                          // the script will take the comments from the following list
        "awesome",                          // again, add as many as you want
        "wow",                              // it will randomly choose one 
        "nice pic"                          // each time it has to put a comment
    ]
}
```

### Note on the emails

I use Gmail to send and receive the emails. If you want to use Gmail, there's something important to do first (from the [Google accounts](https://support.google.com/accounts/answer/6010255) website) :
```
Change your settings to allow less secure apps to access your account.
We don't recommend this option because it might make it easier for someone to break into your account.
If you want to allow access anyway, follow these steps:
    - Go to the "Less secure apps" section in My Account.
    - Next to "Access for less secure apps," select Turn on.
```
(If you can't find where it is exactly, I think [this link](https://myaccount.google.com/security) should work)

As this procedure might not be safe, I recommend not doing it on your main Gmail account, and maybe create another account on the side. Or try to find a less secure webmail provider!

## How to run
This is it! You now have two options : 
1. Run it in the Go playground

   In a terminal, go to the folder where instabot.go is located, then execute :

   `go run instabot.go`

2. Build and install the script as a new command

   In a terminal, write :

   `go install github.com/tducasse/go-instabot.go`

   You now have the `go-instabot` command available from anywhere in your system. Just launch it in a terminal!

## Tips
- Each time you will run the script, it will create a log file, where everything will be recorded.
  If you'd rather not have the logs and check everything directly in your terminal, there's actually an option for that. Run :
  
  `go run instabot.go -nolog`
  
  and it will only be printed on the screen.
- If you want to launch a long session, and you're afraid of closing the terminal, I recommend using the command __screen__.
- If you have a Raspberry Pi, a web server, or anything similar, you can run the script there (again, use screen).
- To maximize your chances of getting new followers, don't spam! If you follow too many people, you will become irrelevant.

  Also, try to use hashtags related to your own account : if you are a portrait photographer and you suddenly start following a thousand #cats related accounts, I doubt it will bring you back a thousand new followers...
  
Good luck getting new followers!

### Donate
If you like my work, feel free to donate!

- BTC : 1PN6TKgBXWmUDRQEmVNvqqPgacgJNndkVk
- ETH : 0x2fFc531893072d17bC63263aC8e8AF9d9E76E751
