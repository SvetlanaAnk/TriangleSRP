package main

import (
	"fmt"
	"log"
	"strings"

	dg "github.com/bwmarrin/discordgo"
)

var (
	commandHandlers = map[string]func(session *dg.Session, interaction *dg.InteractionCreate){
		"add-loss":     addLoss,
		"set-ship-srp": setShipSrp,
		"remove-loss":  removeLoss,
		"update-loss":  updateLoss,
		"srp-paid":     srpPaid,
		"paid":         paid,
		"print-ships":  printShips,
	}
)

func addLoss(session *dg.Session, interaction *dg.InteractionCreate) {
	userIsFc := isUserFc(interaction)
	userName := interaction.Member.User.Username
	optionMap := *generateOptionMap(interaction)
	warning := ""
	var link string
	srp := uint64(1)

	if opt, ok := optionMap["link"]; ok {
		link = opt.StringValue()
	}

	// Verify that the link is valid, and pass it through ToLower() and regex
	parsedLink := regexMatchZkill(strings.ToLower(link))

	if parsedLink == "" {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Invalid Zkill format: %v", link))
		return
	}

	// Check if this loss already exists on the Srp sheet
	loss := *getLossFromLink(parsedLink)
	if loss != (Losses{}) {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Link has already been submitted\n%v", link))
		return
	}

	//Query the Zkill and Eve api's for needed information
	eveLossData := getLossFromApi(parsedLink)

	ship := getDoctrineShip(uint(eveLossData.Victim.ShipTypeId))

	// Check if the ship is a doctrine ship.
	if *ship == (DoctrineShips{}) {
		if !userIsFc {
			log.Println("Ship not doctrine ship")
			sendInteractionResponse(session, interaction, "Ship is not a valid doctrine ship, please ask an FC to override")
			return
		} else {
			warning += "\tShip is not a registered doctrine hull\n"
		}
	} else {
		srp = ship.Srp
	}

	// Check if the ship died in pochven
	if !isPochvenSystem(eveLossData.SolarSystemId) {
		if !userIsFc {
			sendInteractionResponse(session, interaction, "This ship was destroyed outside of Pochven, please ask an FC to override")
			return
		} else {
			warning += "\tShip was not destroyed in Pochven\n"
		}
	}

	if warning != "" {
		warning = "Warning(s):\n" + warning + "Fc has overriden"
	}

	// If a custom user was selected to receive srp, use that instead
	if opt, ok := optionMap["user"]; ok {
		userName = opt.UserValue(session).Username

		if userName == session.State.User.Username {
			sendInteractionResponse(session, interaction, "While I am flattered, I cannot receive Srp since I am a bot.\nPlease select a capsuleer, or at least a fellow bot in Fraternity.")
			return
		}
		if opt.UserValue(session).Bot {
			sendInteractionResponse(session, interaction, "My fellow bots cannot receive Srp.\nPlease select a capsuleer, or at least a Fraternity member.")
			return
		}
	}

	// Get the custom srp amount if relevant, only Fc's can pass in this value
	if opt, ok := optionMap["srp"]; ok {
		srp = uint64(opt.IntValue())
		if !isUserFc(interaction) {
			sendInteractionResponse(session, interaction, "Only an FC can specify a custom Srp amount.")
			return
		}
	}

	//Submit the loss to the database, and report the result to the user
	loss = Losses{UserName: userName, Url: parsedLink, Srp: srp, ShipId: uint(eveLossData.Victim.ShipTypeId)}

	creationResult := db.Create(&loss)

	if creationResult.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error submitting Link. %v", link))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Submitted successfully\nLoss: %s\nAmount: %v million isk\nCapsuleer: %v\n%s", link, srp, userName, warning))
	}
}

func setShipSrp(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction) {
		sendInteractionResponse(session, interaction, "You are not an FC..")
		return
	}
	optionMap := *generateOptionMap(interaction)

	shipID := uint(0)
	srp := uint64(1)
	if opt, ok := optionMap["ship-id"]; ok {
		shipID = uint(opt.IntValue())
	}
	if opt, ok := optionMap["srp"]; ok {
		srp = uint64(opt.IntValue())
	}
	ship := *getDoctrineShip(uint(shipID))
	if ship != (DoctrineShips{}) {
		db.Model(&DoctrineShips{}).Where("id = ?", shipID).Update("srp", srp)
		sendInteractionResponse(session, interaction, fmt.Sprintf("ShipId %d was already present. Srp value has been updated to %d million Isk", shipID, srp))
		return
	}
	shipName := getShipNameFromId(uint(shipID))

	if shipName == "" {
		sendInteractionResponse(session, interaction, fmt.Sprintf("ShipId: %d not valid", shipID))
		return
	}
	ship = DoctrineShips{Ship_ID: shipID, Name: shipName, Srp: srp}
	creationResult := db.Create(&ship)

	if creationResult.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error creating ship %v : %v\n%v", ship.Ship_ID, ship.Name, creationResult.Error))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Doctrine ship %s ID: %d added with an SRP value of %d million Isk", ship.Name, shipID, srp))
	}
}

func removeLoss(session *dg.Session, interaction *dg.InteractionCreate) {
	optionMap := *generateOptionMap(interaction)
	var link string

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
		sendInteractionResponse(session, interaction, fmt.Sprintf("Loss not found\n%v", link))
		return
	}

	if !isUserFc(interaction) && loss.UserName != interaction.Member.User.Username {
		sendInteractionResponse(session, interaction, "Only an FC can delete someone else's loss.")
		return
	}

	result := db.Where("url = ?", parsedLink).Delete(&loss)

	if result.Error == nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Loss has been removed\n%v", link))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error removing loss: %v\n%v", link, result.Error))
	}
}

func updateLoss(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction) {
		sendInteractionResponse(session, interaction, "You are not an FC..")
		return
	}

	srp := uint64(1)
	optionMap := *generateOptionMap(interaction)
	var link string

	if opt, ok := optionMap["link"]; ok {
		link = opt.StringValue()
	}

	if opt, ok := optionMap["srp"]; ok {
		srp = uint64(opt.IntValue())
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
	paid := loss.Paid
	user := loss.UserName

	if opt, ok := optionMap["user"]; ok {
		user = opt.UserValue(session).Username
		if user == session.State.User.Username {
			sendInteractionResponse(session, interaction, "While I am flattered, I cannot receive Srp since I am a bot.\nPlease select a capsuleer, or at least a fellow bot in Fraternity.")
			return
		}
		if opt.UserValue(session).Bot {
			sendInteractionResponse(session, interaction, "My fellow bots cannot receive Srp.\nPlease select a capsuleer, or at least a Fraternity member.")
			return
		}
	}

	if opt, ok := optionMap["paid"]; ok {
		paid = opt.BoolValue()
	}

	result := db.Model(&Losses{}).Where("url = ?", parsedLink).Updates(Losses{Srp: srp, Paid: paid, UserName: user})

	if result.Error == nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Loss of: %v\nHas been updated\nSrp: %v million isk\nPaid: %v\nCapsuleer: %v", link, srp, paid, user))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error removing loss: %v\n%v", link, result.Error))
	}
}

func srpPaid(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction) {
		sendInteractionResponse(session, interaction, "You are not an FC..")
		return
	}
	batchId := uint(0)
	row := db.Table("losses").Select("max(batch)").Row()
	row.Scan(&batchId)

	batchId += 1

	result := db.Model(&Losses{}).Where("paid = ?", false).Updates(&Losses{Paid: true, Batch: batchId})
	if result.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Sql error closing backlog: %v", result.Error))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Srp has been marked as paid\nLosses marked as paid: %d\nBatch Id: %d\nPlease save this batch id in case you need to reverse this action", result.RowsAffected, batchId))
	}
}

func paid(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction) {
		sendInteractionResponse(session, interaction, "You are not an FC..")
		return
	}

	optionMap := *generateOptionMap(interaction)
	var link string

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

	result := db.Model(&Losses{}).Where("url = ?", parsedLink).Update("paid", true)
	if result.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Sql error: %v", result.Error))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Loss has been marked as paid\n%s", link))
	}
}

func printShips(session *dg.Session, interaction *dg.InteractionCreate) {
	var ships []DoctrineShips

	result := db.Find(&ships)
	if result.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Sql error: %v", result.Error))
	} else {
		shipString := generateDoctrineShipString(ships)
		if shipString == "" {
			shipString = "There are currently no registered doctrine ships"
		}
		sendInteractionResponse(session, interaction, shipString)
	}

}

func setchannel(session *dg.Session, interaction *dg.InteractionCreate) {

}

func getsrptotals(session *dg.Session, interaction *dg.InteractionCreate) {

}

func messageCreate(session *dg.Session, message *dg.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}
}
