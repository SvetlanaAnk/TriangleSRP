package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		"add-kill": addKill,
	}
)

func addKill(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

	optionMap := *generateOptionMap(interaction)

	link := ""
	if opt, ok := optionMap["link"]; ok {
		link = opt.StringValue()
	}

	parsedLink := regexMatchZkill(strings.ToLower(link))

	if parsedLink == "" {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Invalid Zkill format. %v", link))
		return
	}

	result := db.Where("url = ?", parsedLink).Find(&Losses{})
	if result.RowsAffected != 0 {
		sendInteractionResponse(session, interaction, "Link has already been submitted.")
		return
	}

	lossData, eveLossData := getLossFromApi(parsedLink)
	if lossData != nil && eveLossData != (EveLoss{}) {
		log.Print(lossData[0].KillmailId)
		log.Print(eveLossData.Victim.ShipTypeId)
	}

	userName := interaction.Member.User.Username
	srp := uint(1)

	if opt, ok := optionMap["user"]; ok {
		userName = opt.UserValue(nil).Username
	}

	if opt, ok := optionMap["srp"]; ok {
		srp = uint(opt.IntValue())
	}

	loss := Losses{UserName: userName, Url: parsedLink, Srp: srp}

	creationResult := db.Create(&loss)

	if creationResult.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error submitting Link. %v", link))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Srp Link: %v\nSubmitted successfully for amount: %v million isk\nFor Capsuleer: %v", link, srp, userName))
	}
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

}
