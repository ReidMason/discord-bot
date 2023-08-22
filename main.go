package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var session *discordgo.Session

const RemoveCommands = true

const UpvoteEmoji = "😂"

type UserProfile struct {
	ID       string
	Username string
	Karma    int
}

func addKarma(userId string, username string, users *[]UserProfile, amount int) {
	for _, user := range *users {
		if user.ID == userId {
			user.Karma += amount
			return
		}
	}

	*users = append(*users, UserProfile{ID: userId, Username: username, Karma: amount})
}

type Command struct {
	Name        string
	Description string
	Type        discordgo.ApplicationCommandType
	Handler     func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func (c *Command) GetApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name,
		Description: c.Description,
		Type:        c.Type,
	}
}

var commands = []Command{
	{
		Name:        "ping",
		Description: "Basic ping command",
		Type:        discordgo.ApplicationCommandType(1),
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			now := time.Now()
			msg, _ := s.InteractionResponse(i.Interaction)
			elapsed := now.Sub(msg.Timestamp)

			// responseContent := fmt.Sprintf("Pong!\nClient: %dms\nWebsocket: %dms", elapsed.Milliseconds(), s.HeartbeatLatency().Milliseconds())
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{
					{
						Title: "Pong!",
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:  "Client",
								Value: fmt.Sprintf("%dms", elapsed.Milliseconds()),
							},
							{
								Name:  "Websocket",
								Value: fmt.Sprintf("%dms", s.HeartbeatLatency().Milliseconds()),
							},
						},
					},
				},
			})
		},
	},
	{
		Name:        "calckarma",
		Description: "Calculate karma score of users",
		Type:        discordgo.ApplicationCommandType(1),
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			channelId := i.ChannelID

			messages := make([]*discordgo.Message, 0)
			messageBatches := 10
			messagesPerBatch := 100
			lastMessageId := ""
			log.Print("Getting messages")
			for i := 0; i < messageBatches; i++ {
				newMessages, err := s.ChannelMessages(channelId, messagesPerBatch, lastMessageId, "", "")
				if err == nil {
					messages = append(messages, newMessages...)
				}
				lastMessageId = newMessages[len(newMessages)-1].ID
			}

			users := make([]UserProfile, 0)

			log.Print("Calculating karma")
			for _, message := range messages {
				reactions := message.Reactions
				for _, reaction := range reactions {
					if reaction.Emoji.Name == UpvoteEmoji {
						addKarma(message.Author.ID, message.Author.Username, &users, 1)
					}
				}
			}

			log.Print("Building fields")
			fields := make([]*discordgo.MessageEmbedField, 0)
			for _, user := range users {
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:  user.Username,
					Value: fmt.Sprintf("Karma: %d", user.Karma),
				})
			}

			log.Print("Sending message")
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &[]*discordgo.MessageEmbed{
					{
						Title:       "Karma count",
						Description: fmt.Sprintf("Last %d messages", len(messages)),
						Fields:      fields,
					},
				},
			})
		},
	},
	{
		Name:        "test",
		Description: "A test command",
		Type:        discordgo.ApplicationCommandType(1),
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:       "Testing",
							Description: "Description testing",
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:  "Field 1",
									Value: "Value 1",
								},
								{
									Name:  "Field 2",
									Value: "Value 2",
								},
							},
						},
					},
				},
			})
		},
	},
}

func main() {
	godotenv.Load(".env")
	accessToken := os.Getenv("TOKEN")

	s, err := discordgo.New("Bot " + accessToken)
	if err != nil {
		log.Println("Failed to start bot", err)
		return
	}

	s.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Register command handlers
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		name := i.ApplicationCommandData().Name
		for _, command := range commands {
			if command.Name == name {
				command.Handler(s, i)
				break
			}
		}
	})

	// Open a websocket connection to Discord and begin listening.
	err = s.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	guilds := s.State.Guilds
	registeredCommands := make([][]*discordgo.ApplicationCommand, len(guilds))
	for i := range guilds {
		registeredCommands[i] = make([]*discordgo.ApplicationCommand, len(commands))
	}

	for i, guild := range guilds {
		log.Printf("Adding commands for %s (%s)...", guild.Name, guild.ID)
		for j, v := range commands {
			cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guild.ID, v.GetApplicationCommand())
			if err != nil {
				log.Printf("Cannot create command '%v' error: %v", v.Name, err)
			}
			registeredCommands[i][j] = cmd
		}
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if RemoveCommands {
		for i, guild := range guilds {
			log.Printf("Removing commands for %s...", guild.Name)
			for _, v := range registeredCommands[i] {
				err := s.ApplicationCommandDelete(s.State.User.ID, guild.ID, v.ID)
				if err != nil {
					log.Printf("Cannot delete command '%v' error: %v", v.Name, err)
				}
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
