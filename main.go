package main

import (
	"discord-bot/internal/commands"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var session *discordgo.Session

const RemoveCommands = true

var botCommands = commands.GetCommands()

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
		for _, command := range botCommands {
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
		registeredCommands[i] = make([]*discordgo.ApplicationCommand, len(botCommands))
	}

	for i, guild := range guilds {
		log.Printf("Adding commands for %s (%s)...", guild.Name, guild.ID)
		for j, v := range botCommands {
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
