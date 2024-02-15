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
}

func init() {
	dg_session.AddHandler(func(s *dg.Session, i *dg.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	dg_session.AddHandler(messageCreate)
}

func main() {
	dg_session.AddHandler(func(s *dg.Session, r *dg.Ready) {
		log.Printf("Logged in as %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	dg_session.Identify.Intents = dg.IntentsGuildMessages

	err := dg_session.Open()

	if err != nil {
		log.Fatalf("Cannot open the session :%v", err)
	}

	log.Println("Adding commands...")

	//registeredCommands := make([]*dg.ApplicationCommand, len(commands))

	//dg_session.ApplicationCommandBulkOverwrite("1205737918556147722", GUILD_ID, commands)
	log.Println("Commands Added")
	// for i, v := range commands {
	// 	cmd, err := dg_session.ApplicationCommand(dg_session.State.User.ID, GUILD_ID, v)
	// 	if err != nil {
	// 		log.Panicf("Cannot create '%v' command: %v", v.Name, err)
	// 	}
	// 	registeredCommands[i] = cmd
	// }

	defer dg_session.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press CTRL+C to exit")
	<-stop

	//log.Println("Removing Commands...")

	// for _, v := range registeredCommands {
	// 	err := dg_session.ApplicationCommandDelete(dg_session.State.User.ID, GUILD_ID, v.ID)
	// 	if err != nil {
	// 		log.Panicf("Cannot delete '%v' command :%v", v.Name, err)
	// 	}
	// }

	log.Println("Shutting Down")
}
