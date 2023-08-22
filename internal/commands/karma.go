package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

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

func karmaCommand() Command {
	return Command{
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
	}
}
