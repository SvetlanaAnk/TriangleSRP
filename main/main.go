package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	DISCORD_TOKEN string
	GUILD_ID      string
	dg_session    *discordgo.Session
)

func init() {
	DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")
	GUILD_ID = os.Getenv("GUILD_ID")

	var err error
	dg_session, err = discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		log.Fatalf("Invalid bot paramters: %v", err)
	}
}

func init() {
	dg_session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	dg_session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	dg_session.AddHandler(messageCreate)
	dg_session.Identify.Intents = discordgo.IntentsGuildMessages

	err := dg_session.Open()

	if err != nil {
		log.Fatalf("Cannot open the session :%v", err)
	}

	log.Println("Adding commands...")

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	for i, v := range commands {
		cmd, err := dg_session.ApplicationCommandCreate(dg_session.State.User.ID, GUILD_ID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer dg_session.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press CTRL+C to exit")
	<-stop

	log.Println("Removing Commands...")

	for _, v := range registeredCommands {
		err := dg_session.ApplicationCommandDelete(dg_session.State.User.ID, GUILD_ID, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command :%v", v.Name, err)
		}
	}

	log.Println("Shutting Down")
}
