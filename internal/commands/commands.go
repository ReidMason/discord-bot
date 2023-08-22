package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

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

func GetCommands() []Command {
	return []Command{
		pingCommand(),
		testCommand(),
		karmaCommand(),
	}
}

func pingCommand() Command {
	return Command{
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
	}
}

func testCommand() Command {
	return Command{
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
	}
}
