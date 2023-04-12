package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	EnvSlackWebhook             = "SLACK_WEBHOOK"
	EnvSlackIcon                = "SLACK_ICON"
	EnvSlackIconEmoji           = "SLACK_ICON_EMOJI"
	EnvSlackChannel             = "SLACK_CHANNEL"
	EnvSlackTitle               = "SLACK_TITLE"
	EnvSlackMessage             = "SLACK_MESSAGE"
	EnvSlackColor               = "SLACK_COLOR"
	EnvSlackUserName            = "SLACK_USERNAME"
	EnvSlackFooter              = "SLACK_FOOTER"
	EnvGithubActor              = "GITHUB_ACTOR"
	EnvGithubLastCommitAuthor   = "GITHUB_LAST_COMMIT_AUTHOR"
	EnvGithubLastCommitMessage  = "GITHUB_LAST_COMMIT_MESSAGE"
	EnvGithubLastCommitLongSHA  = "GITHUB_LAST_COMMIT_LONG_SHA"
	EnvGithubLastCommitShortSHA = "GITHUB_LAST_COMMIT_SHORT_SHA"
	EnvSiteName                 = "SITE_NAME"
	EnvHostName                 = "HOST_NAME"
	EnvMinimal                  = "MSG_MINIMAL"
	EnvSlackLinkNames           = "SLACK_LINK_NAMES"
	EnvTestDuration             = "TEST_DURATION"
	EnvTestStart                = "TEST_START"
	EnvTestSummary              = "TEST_SUMMARY"
)

type Webhook struct {
	Text        string       `json:"text,omitempty"`
	UserName    string       `json:"username,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	LinkNames   string       `json:"link_names,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback   string  `json:"fallback"`
	Pretext    string  `json:"pretext,omitempty"`
	Color      string  `json:"color,omitempty"`
	AuthorName string  `json:"author_name,omitempty"`
	AuthorLink string  `json:"author_link,omitempty"`
	AuthorIcon string  `json:"author_icon,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	AtSomeone  string  `json:"at_someone,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
}

type Field struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

func main() {
	endpoint := os.Getenv(EnvSlackWebhook)
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "URL is required")
		os.Exit(1)
	}
	text := os.Getenv(EnvSlackMessage)
	if text == "" {
		fmt.Fprintln(os.Stderr, "Message is required")
		os.Exit(1)
	}
	if strings.HasPrefix(os.Getenv("GITHUB_WORKFLOW"), ".github") {
		os.Setenv("GITHUB_WORKFLOW", "Link to action run")
	}

	// long_sha := os.Getenv("GITHUB_SHA")
	// commit_sha := long_sha[0:6]

	minimal := os.Getenv(EnvMinimal)
	fields := []Field{}
	if minimal == "true" {
		mainFields := []Field{
			{
				Title: os.Getenv(EnvSlackTitle),
				Value: envOr(EnvTestSummary, "EOM"),
				Short: false,
			},
		}
		fields = append(mainFields, fields...)
	} else if minimal != "" {
		requiredFields := strings.Split(minimal, ",")
		mainFields := []Field{
			{
				Title: os.Getenv(EnvSlackTitle),
				Value: envOr(EnvTestSummary, "EOM"),
				Short: false,
			},
		}
		for _, requiredField := range requiredFields {
			switch strings.ToLower(requiredField) {
			case "duration":
				field := []Field{
					{
						Title: "Duration",
						Value: os.Getenv("TEST_DURATION"),
						Short: true,
					},
				}
				mainFields = append(field, mainFields...)
			case "start":
				field := []Field{
					{
						Title: "Start",
						Value: os.Getenv("TEST_START"),
						Short: true,
					},
				}
				mainFields = append(field, mainFields...)
			case "actions url":
				field := []Field{
					{
						Title: "Actions URL",
						Value: "<" + os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/commit/" + os.Getenv("GITHUB_SHA") + "/checks|" + os.Getenv("GITHUB_WORKFLOW") + ">",
						Short: true,
					},
				}
				mainFields = append(field, mainFields...)
			case "commit":
				field := []Field{
					{
						Title: "Last Commit Message: " + envOr(EnvGithubLastCommitMessage, "Commit"),
						Value: "<" + os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/commit/" + os.Getenv("GITHUB_LAST_COMMIT_LONG_SHA") + "|" + os.Getenv("GITHUB_LAST_COMMIT_SHORT_SHA") + ">",
						Short: true,
					},
				}
				mainFields = append(field, mainFields...)
			}
		}
		fields = append(mainFields, fields...)
	} else {
		mainFields := []Field{
			{
				Title: "Duration",
				Value: os.Getenv("TEST_DURATION"),
				Short: true,
			}, {
				Title: "Start",
				Value: os.Getenv("TEST_START"),
				Short: true,
			},
			{
				Title: "Actions URL",
				Value: "<" + os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/commit/" + os.Getenv("GITHUB_SHA") + "/checks|" + os.Getenv("GITHUB_WORKFLOW") + ">",
				Short: true,
			},
			{
				Title: "Last Commit Message: " + envOr(EnvGithubLastCommitMessage, "Commit"),
				Value: "<" + os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/commit/" + os.Getenv("GITHUB_LAST_COMMIT_LONG_SHA") + "|" + os.Getenv("GITHUB_LAST_COMMIT_SHORT_SHA") + ">",
				Short: true,
			},
			{
				Title: os.Getenv(EnvSlackTitle),
				Value: envOr(EnvTestSummary, "EOM"),
				Short: false,
			},
		}
		fields = append(mainFields, fields...)
	}

	hostName := os.Getenv(EnvHostName)
	if hostName != "" {
		newfields := []Field{
			{
				Title: os.Getenv("SITE_TITLE"),
				Value: os.Getenv(EnvSiteName),
				Short: true,
			},
			{
				Title: os.Getenv("HOST_TITLE"),
				Value: os.Getenv(EnvHostName),
				Short: true,
			},
		}
		fields = append(newfields, fields...)
	}

	color := ""
	switch os.Getenv(EnvSlackColor) {
	case "success":
		color = "good"
	case "cancelled":
		color = "#808080"
	case "failure":
		color = "danger"
	default:
		color = envOr(EnvSlackColor, "good")
	}

	msg := Webhook{
		UserName:  os.Getenv(EnvSlackUserName),
		IconURL:   os.Getenv(EnvSlackIcon),
		IconEmoji: os.Getenv(EnvSlackIconEmoji),
		Channel:   os.Getenv(EnvSlackChannel),
		LinkNames: os.Getenv(EnvSlackLinkNames),
		Attachments: []Attachment{
			{
				Fallback:   envOr(EnvTestSummary, "GITHUB_ACTION="+os.Getenv("GITHUB_ACTION")+" \n GITHUB_ACTOR="+os.Getenv("GITHUB_ACTOR")+" \n GITHUB_EVENT_NAME="+os.Getenv("GITHUB_EVENT_NAME")+" \n GITHUB_REF="+os.Getenv("GITHUB_REF")+" \n GITHUB_REPOSITORY="+os.Getenv("GITHUB_REPOSITORY")+" \n GITHUB_WORKFLOW="+os.Getenv("GITHUB_WORKFLOW")),
				Color:      color,
				AuthorName: "Last Commit Author: " + envOr(EnvGithubLastCommitAuthor, ""),
				AuthorLink: os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv(EnvGithubLastCommitAuthor),
				AuthorIcon: os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv(EnvGithubLastCommitAuthor) + ".png?size=32",
				AtSomeone:  "<@" + os.Getenv("SLACK_AT_USERID") + ">",
				Footer:     envOr(EnvSlackFooter, "<https://github.com/poper-inc/action-slack-notify|Powered By poper-inc's gitHub actions library>"),
				Fields:     fields,
			},
		},
	}

	if err := send(endpoint, msg); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %s\n", err)
		os.Exit(2)
	}
}

func envOr(name, def string) string {
	if d, ok := os.LookupEnv(name); ok {
		return d
	}
	return def
}

func send(endpoint string, msg Webhook) error {
	enc, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(enc)
	res, err := http.Post(endpoint, "application/json", b)
	if err != nil {
		return err
	}

	if res.StatusCode >= 299 {
		return fmt.Errorf("error on message: %s", res.Status)
	}
	fmt.Println(res.Status)
	return nil
}
