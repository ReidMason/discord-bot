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

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Basic ping command",
		Type:        discordgo.ApplicationCommandType(1),
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		now := time.Now()
		msg, _ := s.InteractionResponse(i.Interaction)
		elapsed := now.Sub(msg.Timestamp)

		responseContent := fmt.Sprintf("Pong!\nClient: %dms\nWebsocket: %dms", elapsed.Milliseconds(), s.HeartbeatLatency().Milliseconds())
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &responseContent,
		})
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
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
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
			cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guild.ID, v)
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
