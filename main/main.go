package main

import (
	"log"
	"os"
	"os/signal"

	dg "github.com/bwmarrin/discordgo"
)

var (
	DISCORD_TOKEN string
	GUILD_ID      string
	dg_session    *dg.Session
)

func init() {
	DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")
	GUILD_ID = os.Getenv("GUILD_ID")

	var err error
	dg_session, err = dg.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		log.Fatalf("Invalid bot paramters: %v", err)
	}

	dg_session.AddHandler(func(s *dg.Session, i *dg.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func registerCommand(command *dg.ApplicationCommand, channel chan *dg.ApplicationCommand) {
	cmd, err := dg_session.ApplicationCommandCreate(dg_session.State.User.ID, GUILD_ID, command)
	if err != nil {
		log.Panicf("Cannot create '%v' command %v", command.Name, err)
	}
	channel <- cmd
}

func main() {
	dg_session.AddHandler(func(s *dg.Session, r *dg.Ready) {
		log.Printf("Logged in as %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	//dg_session.AddHandler(messageCreate)
	dg_session.Identify.Intents = dg.IntentsGuildMessages

	err := dg_session.Open()

	if err != nil {
		log.Fatalf("Cannot open the session :%v", err)
	}

	log.Println("Adding commands...")

	registeredCommands := make([]*dg.ApplicationCommand, len(commands))
	channel := make(chan *dg.ApplicationCommand)
	for _, command := range commands {
		go registerCommand(command, channel)
	}

	for i := range commands {
		registeredCommands[i] = <-channel
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
