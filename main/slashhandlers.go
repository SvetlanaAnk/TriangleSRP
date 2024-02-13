package main

import (
	"fmt"
	"strings"

	dg "github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

var (
	commandHandlers = map[string]func(session *dg.Session, interaction *dg.InteractionCreate){
		"add-loss":       addLoss,
		"set-ship-srp":   setShipSrp,
		"remove-loss":    removeLoss,
		"update-loss":    updateLoss,
		"srp-paid":       srpPaid,
		"paid":           paid,
		"print-ships":    printShips,
		"srp-totals":     srpTotals,
		"user-srp-total": userSrpTotal,
		"remove-ship":    removeDoctrineShip,
		"set-channel":    setSrpChannel,
		"mark-user-paid": markUserPaid,
	}
)

func addLoss(session *dg.Session, interaction *dg.InteractionCreate) {
	var link string
	options := *generateOptionMap(interaction)
	userName := interaction.Member.User.Username
	userIsFc := isUserFc(interaction.Member)
	customSrp := uint64(0)

	// If a custom user was selected to receive srp, use that instead
	if opt, ok := options["user"]; ok {
		userName = opt.UserValue(session).Username

		if userName == session.State.User.Username {
			sendInteractionResponse(session, interaction, "While I am flattered, I cannot receive Srp since I am a bot.\nPlease select a capsuleer, or at least a fellow bot in Fraternity.")
		}
		if opt.UserValue(session).Bot {
			sendInteractionResponse(session, interaction, "My fellow bots cannot receive Srp.\nPlease select a capsuleer, or at least a Fraternity member.")
		}
	}

	if opt, ok := options["link"]; ok {
		link = opt.StringValue()
	}

	// If a custom Srp amount is passed, use that
	if opt, ok := options["srp"]; ok {
		customSrp = uint64(opt.IntValue())
		if !userIsFc {
			sendInteractionResponse(session, interaction, "Only an FC can specify a custom Srp amount.")
		}
	}

	result := addKill(userName, link, userIsFc, customSrp)
	sendInteractionResponse(session, interaction, result)
}

func setShipSrp(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member) {
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
		result := db.Model(&DoctrineShips{}).Where("ship_id = ?", shipID).Update("srp", srp)
		if result.Error == nil && result.RowsAffected == 1 {
			sendInteractionResponse(session, interaction, fmt.Sprintf("ShipId %d was already present. Srp value has been updated to %d million Isk", shipID, srp))
			return
		} else {
			sendInteractionResponse(session, interaction, fmt.Sprintf("Sql Error Updating Ship: %v", result.Error))
			return
		}
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

	if !isUserFc(interaction.Member) && loss.UserName != interaction.Member.User.Username {
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
	if !isUserFc(interaction.Member) {
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
	if !isUserFc(interaction.Member) {
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
	if !isUserFc(interaction.Member) {
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

func markUserPaid(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member) {
		sendInteractionResponse(session, interaction, "You are not an FC..")
		return
	}

	optionMap := *generateOptionMap(interaction)
	var userName string

	if opt, ok := optionMap["user"]; ok {
		userName = opt.UserValue(session).Username
	}

	res := db.Model(&Losses{}).Where("user_name = ?", userName).Update("paid", true)
	if res.RowsAffected == 0 {
		sendInteractionResponse(session, interaction, fmt.Sprintf("No losses found for user: %s", userName))
	} else {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Member: %s's losses have been marked as paid.\nNumber paid: %d", userName, res.RowsAffected))
	}
}

func printShips(session *dg.Session, interaction *dg.InteractionCreate) {
	var ships []DoctrineShips

	result := db.Find(&ships)
	if result.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Sql error querying ships: %v", result.Error))
	} else {
		shipString := generateDoctrineShipString(ships)
		if shipString == "" {
			shipString = "There are currently no registered doctrine ships"
		}
		sendInteractionResponse(session, interaction, shipString)
	}
}

func srpTotals(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member) {
		sendInteractionResponse(session, interaction, "You are not an FC...")
		return
	}
	optionMap := *generateOptionMap(interaction)
	var userName string

	if opt, ok := optionMap["user"]; ok {
		userName = opt.UserValue(session).Username
	}

	var losses []Losses

	var result *gorm.DB

	if userName == "" {
		result = db.Where("paid = ?", false).Find(&losses)
	} else {
		result = db.Where("user_name = ? AND paid = ?", userName, false).Find(&losses)
	}

	if result.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error while querying losses: %v", result.Error))
		return
	}
	if len(losses) == 0 {
		sendInteractionResponse(session, interaction, "No unpaid losses found.")
		return
	}

	printZkill := false
	if opt, ok := optionMap["include-zkill"]; ok {
		printZkill = opt.BoolValue()
	}

	printWarnings := false
	if opt, ok := optionMap["include-errors"]; ok {
		printWarnings = opt.BoolValue()
	}

	lossTotals := generateSrpTotalString(losses, printZkill, printWarnings)
	sendInteractionResponse(session, interaction, fmt.Sprintf("SRP Totals Per Character\n%s", lossTotals))
}

func userSrpTotal(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member) {
		sendInteractionResponse(session, interaction, "You are not an FC...")
		return
	}

	optionMap := *generateOptionMap(interaction)

	var userName string

	if opt, ok := optionMap["user"]; ok {
		userName = opt.UserValue(session).Username
	} else {
		userName = interaction.Member.User.Username
	}

	var losses []Losses
	result := db.Where("user_name = ? AND paid = ?", userName, false).Find(&losses)
	if result.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("SQL Error while querying losses: %v", result.Error))
		return
	}
	if len(losses) == 0 {
		sendInteractionResponse(session, interaction, "No unpaid losses found.")
		return
	}

	lossTotals := generateSrpTotalForUser(losses)
	sendInteractionResponse(session, interaction, lossTotals)
}

func removeDoctrineShip(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member) {
		sendInteractionResponse(session, interaction, "You are not an FC...")
		return
	}

	optionMap := *generateOptionMap(interaction)

	shipId := uint(0)

	if opt, ok := optionMap["ship-id"]; ok {
		shipId = uint(opt.IntValue())
	}
	ship := DoctrineShips{}
	db.Where("ship_id = ?", shipId).First(&ship)

	result := db.Delete(&ship)

	if result.Error == nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Ship %s ID: %d removed from doctrine ships", ship.Name, ship.Ship_ID))
	} else {
		sendInteractionResponse(session, interaction, "ShipId not found:")
	}
}

func setSrpChannel(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member) {
		sendInteractionResponse(session, interaction, "You are not an FC...")
		return
	}
	config := ServerConfiguration{}

	db.Where("guild_id = ?", interaction.GuildID).First(&config)

	var res *gorm.DB

	if config == (ServerConfiguration{}) {
		config.GuildId = interaction.GuildID
		config.SrpChannel = interaction.ChannelID
		res = db.Create(&config)
	} else {
		res = db.Model(&ServerConfiguration{}).Where("guild_id = ?", config.GuildId).Updates(ServerConfiguration{SrpChannel: interaction.ChannelID})
	}

	if res.Error != nil {
		sendInteractionResponse(session, interaction, fmt.Sprintf("Sql Query updating Srp channel: %v", res.Error))
		return
	}

	SRP_CHANNEL_MAP[interaction.GuildID] = interaction.ChannelID

	sendInteractionResponse(session, interaction, "Srp channel set successfully")
}

func messageCreate(session *dg.Session, message *dg.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	srpChannelId := ""

	if arg, ok := SRP_CHANNEL_MAP[message.GuildID]; ok {
		srpChannelId = arg
	}

	if srpChannelId != message.ChannelID || srpChannelId == "" {
		return
	}

	userName := message.Author.Username
	member, err := session.GuildMembersSearch(message.GuildID, message.Author.Username, 1)
	if err != nil {
		session.ChannelMessageSendReply(srpChannelId, fmt.Sprintf("Error querying server member: %v", err), message.Reference())
	}
	userIsFc := isUserFc(member[0])
	result := addKill(userName, message.Content, userIsFc, 0)
	session.ChannelMessageSendReply(srpChannelId, result, message.Reference())
}
