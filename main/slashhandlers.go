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
		"user-srp-total": userSrpTotal,
		"remove-ship":    removeDoctrineShip,
		"set-channel":    setSrpChannel,
		"mark-user-paid": markUserPaid,
		"add-fc":         addFc,
		"srp-totals":     srpTotals,
		"rollback-batch": rollBackBatch,
		"remove-fc":      removeFc,
	}
)

func addLoss(session *dg.Session, interaction *dg.InteractionCreate) {
	var link string
	options := *generateOptionMap(interaction)
	nickName := ""
	userId := interaction.Member.User.ID
	userIsFc := isUserFc(interaction.Member.User)
	customSrp := uint64(0)

	// If a custom user was selected to receive srp, use that instead
	if opt, ok := options["user"]; ok {
		user := opt.UserValue(session)
		nickName = getNicknameFromUser(session, user)
		userId = user.ID

		if nickName == session.State.User.Username {
			sendSimpleEmbedResponse(session, interaction, "Nice Try", "While I am flattered, I cannot receive Srp since I am a bot.\nPlease select a capsuleer, or at least a fellow bot in Fraternity")
		}
		if opt.UserValue(session).Bot {
			sendSimpleEmbedResponse(session, interaction, "Nice try", "My fellow bots cannot receive Srp")
		}
	} else {
		nickName = getNicknameFromUser(session, interaction.Member.User)
	}

	if opt, ok := options["link"]; ok {
		link = opt.StringValue()
	}

	// If a custom Srp amount is passed, use that
	if opt, ok := options["srp"]; ok {
		customSrp = uint64(opt.IntValue())
		if !userIsFc {
			sendSimpleEmbedResponse(session, interaction, "You cannot do that", "Only an FC can specify a custom Srp amount")
		}
	}

	result := addKill(nickName, userId, link, userIsFc, customSrp)
	sendInteractionResponse(session, interaction, result)
}

func setShipSrp(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
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
			sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Doctrine Ship: %s Id: %d was already present\nSrp value updated to %d million Isk", ship.Name, shipID, srp), "Srp Updated")
			return
		} else {
			sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Sql Error Updating Ship: %v", result.Error), "Sql Error")
			return
		}
	}

	shipName := getShipNameFromId(uint(shipID))

	if shipName == "" {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Ship Id: %d not valid", shipID), "Invalid Id")
		return
	}
	ship = DoctrineShips{Ship_ID: shipID, Name: shipName, Srp: srp}
	creationResult := db.Create(&ship)

	if creationResult.Error != nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("SQL Error creating ship %v : %s\n%v", ship.Ship_ID, ship.Name, creationResult.Error), "Sql Error")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Doctrine Ship %s ID: %d submitted\nSRP value: %d million Isk", ship.Name, shipID, srp), "Doctrine Ship Submitted")
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
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Invalid Zkill format: %v", link), "Invalid Link")
		return
	}

	loss := *getLossFromLink(parsedLink)
	if loss == (Losses{}) {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Loss not found: %v", link), "Not Found")
		return
	}

	if !isUserFc(interaction.Member.User) && loss.UserId != interaction.Member.User.ID {
		sendSimpleEmbedResponse(session, interaction, "Only an FC can delete someone else's loss.", "Permission Denied")
		return
	}

	result := db.Where("url = ?", parsedLink).Delete(&loss)

	if result.Error == nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Loss has been removed: %v", link), "Loss Removed")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("SQL Error removing loss: %s\n%v", link, result.Error), "Sql Error")
	}
}

func updateLoss(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
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
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Invalid Zkill format: %v", link), "Invalid Lossmail")
		return
	}

	loss := *getLossFromLink(parsedLink)
	if loss == (Losses{}) {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Loss not found: %s", link), "Not Found")
		return
	}

	paid := loss.Paid
	nickName := loss.NickName

	if opt, ok := optionMap["user"]; ok {
		nickName = getNicknameFromUser(session, opt.UserValue(session))
		if nickName == session.State.User.Username {
			sendSimpleEmbedResponse(session, interaction, "While I am flattered, I cannot receive Srp since I am a bot\nPlease select a capsuleer", "Nice try")
			return
		}
		if opt.UserValue(session).Bot {
			sendSimpleEmbedResponse(session, interaction, "My fellow bots cannot receive Srp\nPlease select a capsuleer", "Nice try")
			return
		}
	}

	if opt, ok := optionMap["paid"]; ok {
		paid = opt.BoolValue()
	}

	result := db.Model(&Losses{}).Where("url = ?", parsedLink).Updates(Losses{Srp: srp, Paid: paid, NickName: nickName})

	if result.Error == nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Loss of: %v\nHas been updated\nSrp: %v million isk\nPaid: %v\nCapsuleer: %v", link, srp, paid, nickName), "Loss Updated")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("SQL Error removing loss: %s\n%v", link, result.Error), "Sql Error")
	}
}

func srpPaid(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
		return
	}
	batchId := uint(0)
	row := db.Table("losses").Select("max(batch)").Row()
	row.Scan(&batchId)

	batchId += 1

	result := db.Model(&Losses{}).Where("paid = ?", false).Updates(&Losses{Paid: true, Batch: batchId})
	if result.Error != nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Sql error closing backlog: %v", result.Error), "Sql Error")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Srp has been marked as paid\nLosses marked as paid: %d\nBatch Id: %d", result.RowsAffected, batchId), "Srp Paid!")
	}
}

func paid(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
		return
	}

	optionMap := *generateOptionMap(interaction)
	var link string

	if opt, ok := optionMap["link"]; ok {
		link = opt.StringValue()
	}

	parsedLink := regexMatchZkill(strings.ToLower(link))

	if parsedLink == "" {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Invalid Zkill format: %s", link), "Invalid Lossmail")
		return
	}

	loss := *getLossFromLink(parsedLink)
	if loss == (Losses{}) {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Loss not found: %v", link), "Not Found")
		return
	}

	result := db.Model(&Losses{}).Where("url = ?", parsedLink).Update("paid", true)
	if result.Error != nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Sql error: %v", result.Error), "Sql Error")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Loss has been marked as paid: %s", link), "Loss Paid!")
	}
}

func markUserPaid(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
		return
	}

	optionMap := *generateOptionMap(interaction)
	var nickName string
	var user *dg.User
	if opt, ok := optionMap["user"]; ok {
		user = opt.UserValue(session)
		nickName = getNicknameFromUser(session, user)
	}

	res := db.Model(&Losses{}).Where("user_id = ?", user.ID).Update("paid", true)
	if res.RowsAffected == 0 {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("No losses found for user: %s", nickName), "None Found")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Member: %s's losses have been marked as paid.\nNumber paid: %d", nickName, res.RowsAffected), "Srp Paid")
	}
}

func printShips(session *dg.Session, interaction *dg.InteractionCreate) {
	var ships []DoctrineShips

	result := db.Find(&ships)
	if result.Error != nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Sql error querying ships: %v", result.Error), "Sql Error")
	} else {
		shipString := generateDoctrineShipString(ships)
		if shipString == "" {
			sendSimpleEmbedResponse(session, interaction, "There are no registered doctrine ships", "No Ships found!")
		} else {
			sendSimpleEmbedResponse(session, interaction, shipString, "Doctrine Ships")
		}
	}
}

func srpTotals(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
		return
	}
	optionMap := *generateOptionMap(interaction)
	var user *dg.User = nil

	if opt, ok := optionMap["user"]; ok {
		user = opt.UserValue(session)
	}

	var losses []Losses

	var result *gorm.DB

	if user == nil {
		result = db.Where("paid = ?", false).Find(&losses)
	} else {
		result = db.Where("user_id = ? AND paid = ?", user.ID, false).Find(&losses)
	}
	if result.Error != nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("SQL Error while querying losses: %v", result.Error), "Sql Error")
		return
	}
	if len(losses) == 0 {
		sendSimpleEmbedResponse(session, interaction, "No unpaid losses found", "None Found")
		return
	}

	printZkill := true
	if opt, ok := optionMap["include-zkill"]; ok {
		printZkill = opt.BoolValue()
	}

	printWarnings := true
	if opt, ok := optionMap["include-errors"]; ok {
		printWarnings = opt.BoolValue()
	}
	lossTotals := generateSrpTotalString(losses, printZkill, printWarnings)
	sendSimpleEmbedResponse(session, interaction, lossTotals, "Srp Totals Per Character")
}

func userSrpTotal(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
		return
	}

	optionMap := *generateOptionMap(interaction)

	var user *dg.User
	if opt, ok := optionMap["user"]; ok {
		user = opt.UserValue(session)
	} else {
		user = interaction.Member.User
	}

	var losses []Losses
	result := db.Where("user_id = ? AND paid = ?", user.ID, false).Find(&losses)
	if result.Error != nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("SQL Error while querying losses: %v", result.Error), "Sql Error")
		return
	}
	if len(losses) == 0 {
		sendSimpleEmbedResponse(session, interaction, "No unpaid losses found", "None Found")
		return
	}

	lossTotals := generateSrpTotalForUser(losses)
	sendInteractionResponse(session, interaction, lossTotals)
}

func removeDoctrineShip(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
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
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Ship %s ID: %d removed from doctrine ships", ship.Name, ship.Ship_ID), "Ship Removed")
	} else {
		sendSimpleEmbedResponse(session, interaction, "ShipId not found:", "Not Found")
	}
}

func setSrpChannel(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
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
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Sql Query updating Srp channel: %v", res.Error), "Sql Error")
		return
	}

	SRP_CHANNEL_MAP[interaction.GuildID] = interaction.ChannelID

	sendSimpleEmbedResponse(session, interaction, "Srp channel set successfully", "Channel Set")
}

func addFc(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
		return
	}
	options := *generateOptionMap(interaction)
	var user *dg.User
	if opt, ok := options["user"]; ok {
		user = opt.UserValue(session)
	}

	if isUserFc(user) {
		sendSimpleEmbedResponse(session, interaction, "User is already an fc", "Unecessary")
		return
	}
	admin := Administrators{UserId: user.ID, UserName: user.Username, IsSuperAdmin: false}
	res := db.Create(&admin)

	if res.Error == nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("User: %s is now an Fc", admin.UserName), "Fc Registered")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Sql Error adding fc: %v", res.Error), "Sql Error")
	}
}

func removeFc(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserSuperAdmin(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "Only an administrator may remove an Fc", "Permission Denied")
		return
	}
	options := *generateOptionMap(interaction)
	var user *dg.User
	if opt, ok := options["user"]; ok {
		user = opt.UserValue(session)
	}

	if !isUserFc(user) {
		sendSimpleEmbedResponse(session, interaction, "User is not an fc", "Unecessary")
		return
	}
	admin := Administrators{UserId: user.ID}
	res := db.Delete(&admin)

	if res.Error == nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("User: %s is no longer an Fc", admin.UserName), "Fc Removed")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Sql Error removing fc: %v", res.Error), "Sql Error")
	}
}

func rollBackBatch(session *dg.Session, interaction *dg.InteractionCreate) {
	if !isUserFc(interaction.Member.User) {
		sendSimpleEmbedResponse(session, interaction, "You are not an FC..", "Permission Denied")
		return
	}

	options := *generateOptionMap(interaction)
	batchId := int64(-1)
	if opt, ok := options["batch-id"]; ok {
		batchId = opt.IntValue()
	}

	result := db.Model(&Losses{}).Where("batch = ?", batchId).Updates(&Losses{Paid: false})
	if result.Error != nil {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Sql error closing backlog: %v", result.Error), "Sql Error")
	} else {
		sendSimpleEmbedResponse(session, interaction, fmt.Sprintf("Srp has been marked as paid\nLosses marked as paid: %d\nBatch Id: %d", result.RowsAffected, batchId), "Srp Paid!")
	}
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

	if regexMatchZkill(message.Content) == "" {
		return
	}

	member, err := session.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		session.ChannelMessageSendReply(srpChannelId, fmt.Sprintf("Error querying server member: %v", err), message.Reference())
	}

	userIsFc := isUserFc(member.User)
	result := addKill(getNicknameFromUser(session, message.Author), message.Author.ID, message.Content, userIsFc, 0)
	session.ChannelMessageSendReply(srpChannelId, result, message.Reference())
}
