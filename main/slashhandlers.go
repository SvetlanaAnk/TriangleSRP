package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		"add-loss":    addLoss,
		"add-ship":    addDoctrineShip,
		"remove-loss": removeLoss,
	}
)

func addLoss(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	userIsFc := isUserFc(interaction)
	userName := interaction.Member.User.Username
	optionMap := *generateOptionMap(interaction)
	warning := ""
	link := ""
	srp := uint64(1)

	if opt, ok := optionMap["link"]; ok {
		link = opt.StringValue()
	}

	parsedLink := regexMatchZkill(strings.ToLower(link))

	if parsedLink == "" {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Invalid Zkill format. %v", link))
		return
	}

	loss := *getLossFromLink(parsedLink)
	if loss != (Losses{}) {
		sendInteractionResponse(session, interaction, "Link has already been submitted.")
		return
	}

	lossData, eveLossData := getLossFromApi(parsedLink)
	if lossData != nil && eveLossData != (EveLoss{}) {
		log.Print(lossData[0].KillmailId)
		log.Print(eveLossData.Victim.ShipTypeId)
	}

	if *getDoctrineShip(uint(eveLossData.Victim.ShipTypeId)) != (DoctrineShips{}) {
		if !userIsFc {
			sendInteractionResponse(session, interaction, "Ship is not a valid doctrine ship, please ask an FC to override.")
			return
		} else {
			warning += "Warning: Ship is not a registered doctrine hull.\nFc has overriden this check.\n"
		}
	}

	if !isPochvenSystem(eveLossData.SolarSystemId) {
		if !userIsFc {
			sendInteractionResponse(session, interaction, "This ship was destroyed outside of Pochven, please ask an FC to override.")
			return
		} else {
			warning += "Warning: Ship was not destroyed in Pochven.\nFc has overriden this check."
		}
	}

	ship := getDoctrineShip(loss.ShipId)

	if *ship != (DoctrineShips{}) {
		srp = ship.Srp
	}

	if opt, ok := optionMap["user"]; ok {
		userName = opt.UserValue(session).Username
	}

	if opt, ok := optionMap["srp"]; ok {
		srp = uint64(opt.IntValue())
		if !isUserFc(interaction) {
			sendInteractionResponse(session, interaction, "Only an FC can specify a custom SRP amount.")
			return
		}
	}

	loss = Losses{UserName: userName, Url: parsedLink, Srp: srp, ShipId: uint(eveLossData.Victim.ShipTypeId)}

	creationResult := db.Create(&loss)

	if creationResult.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error submitting Link. %v", link))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Submitted successfully for amount: %v million isk\nFor Capsuleer: %v\n%s", srp, userName, warning))
	}
}

func addDoctrineShip(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if !isUserFc(interaction) {
		sendInteractionResponse(session, interaction, "You are not an FC..")
		return
	}
	optionMap := *generateOptionMap(interaction)

	shipID := uint32(0)
	srp := uint64(1)
	if opt, ok := optionMap["ship-id"]; ok {
		shipID = uint32(opt.IntValue())
	}
	if opt, ok := optionMap["srp"]; ok {
		srp = uint64(opt.IntValue())
	}
	ship := *getDoctrineShip(uint(shipID))
	if ship != (DoctrineShips{}) {
		db.Model(&DoctrineShips{}).Where("id = ?", shipID).Update("srp", srp)
		sendInteractionResponse(session, interaction, fmt.Sprintf("Ship %d was already present. Srp value has been updated to %d million Isk", shipID, srp))
		return
	}

	ship = DoctrineShips{ID: shipID, Name: getShipNameFromId(uint(shipID)), Srp: srp}
	creationResult := db.Create(&ship)

	if creationResult.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error creating ship %v : %v\n%v", ship.ID, ship.Name, creationResult.Error))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Doctrine ship %v : %s added with an SRP value of %d million Isk", shipID, ship.Name, srp))
	}
}

func removeLoss(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
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

	loss := *getLossFromLink(parsedLink)
	if loss == (Losses{}) {
		sendInteractionResponse(session, interaction, "Loss not found.")
		return
	}

	if !isUserFc(interaction) && loss.UserName != interaction.Member.User.Username {
		sendInteractionResponse(session, interaction, "Only an FC can delete someone else's loss.")
		return
	}

	result := db.Where("url = ?", parsedLink).Delete(&loss)

	if result.Error == nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Loss with link of: %v\nHas been removed", link))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error removing loss: %v\n%v", link, result.Error))
	}

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

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

}
