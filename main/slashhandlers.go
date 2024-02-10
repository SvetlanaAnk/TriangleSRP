package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	commandHandlers = map[string]callbackFunction{
		"add-kill": addKill,
	}
)

type callbackFunction func(session *discordgo.Session, interaction *discordgo.InteractionCreate)

func addKill(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Yay it works",
		},
	})
}

func deleteKill(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

}

func setkillsrp(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

}

func backlogpaid(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

}

func setsrprate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

}

func setchannel(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

}

func getsrptotals(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

}

func setadminuser(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

}

func removeadminuser(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	match := regexmatchzkill(message.Content)

	if match {
		log.Println("Match!")
	}

}
