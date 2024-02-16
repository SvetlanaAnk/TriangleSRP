package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"

	dg "github.com/bwmarrin/discordgo"
)

// func sendInteractionResponse(session *dg.Session, interaction *dg.InteractionCreate, message string) {
// 	session.InteractionRespond(interaction.Interaction, &dg.InteractionResponse{
// 		Type: dg.InteractionResponseChannelMessageWithSource,
// 		Data: &dg.InteractionResponseData{
// 			Content: message,
// 		},
// 	})
// }

func sendSimpleEmbedResponse(session *dg.Session, interaction *dg.InteractionCreate, message string, title string) {
	embed := []*dg.MessageEmbed{
		{
			Title: title,
			Fields: []*dg.MessageEmbedField{
				{
					Value: message,
				},
			},
		},
	}

	session.InteractionRespond(interaction.Interaction, &dg.InteractionResponse{
		Type: dg.InteractionResponseChannelMessageWithSource,
		Data: &dg.InteractionResponseData{
			Embeds: embed,
		},
	})
}

func sendEmbedResponse(session *dg.Session, interaction *dg.InteractionCreate, embed []*dg.MessageEmbed) {
	session.InteractionRespond(interaction.Interaction, &dg.InteractionResponse{
		Type: dg.InteractionResponseChannelMessageWithSource,
		Data: &dg.InteractionResponseData{
			Embeds: embed,
		},
	})
}

func generateSimpleEmbed(title string, description string, name string, value string) *dg.MessageEmbed {
	embed := &dg.MessageEmbed{
		Title:       title,
		Description: description,
		Fields: []*dg.MessageEmbedField{
			{
				Name:  name,
				Value: value,
			},
		},
	}
	return embed
}

func addKill(nickName string, userID string, link string, userIsFc bool, customSrp uint64) *dg.MessageEmbed {
	warning := ""
	shortenedWarning := ""
	srp := uint64(1)

	// Verify that the link is valid, and pass it through ToLower() and regex
	parsedLink := regexMatchZkill(strings.ToLower(link))

	if parsedLink == "" {
		return generateSimpleEmbed("❔ Invalid Link ❔", "The Zkill link you have submitted is invalid", "Link", link)
	}

	// Check if this loss already exists on the Srp sheet
	loss := *getLossFromLink(parsedLink)
	if loss != (Losses{}) {
		return generateSimpleEmbed("❌ Invalid Link ❌", "The Zkill link you have submitted already exists", "Link", link)
	}

	//Query the Zkill and Eve api's for needed information
	eveLossData := getLossFromApi(parsedLink)
	if eveLossData == nil {
		return generateSimpleEmbed("❌ Sql Error ❌", "The Zkill link you have submitted caused a Sql Error", "Link", link)
	}
	res := db.Select("kill_mail_id").Where("kill_mail_id = ?", eveLossData.KillmailId).First(&Losses{})

	if res.Error == nil {
		return generateSimpleEmbed("❌ Invalid Link ❌", "The Zkill link you have submitted already exists", "Link", link)
	}
	ship := getDoctrineShip(uint(eveLossData.ShipTypeId))

	// Check if the ship is a doctrine ship.
	if *ship == (DoctrineShips{}) {
		if !userIsFc {
			return generateSimpleEmbed("❌ Invalid Ship ❌", "This ship is not a valid doctrine ship", "Link", link)
		} else {
			warning += "\tShip is not a registered doctrine hull\n"
			shortenedWarning += "Not Doctrine"
			ship.Name = getShipNameFromId(uint(eveLossData.ShipTypeId))
		}
	}

	srp = getDoctrineShipSrp(ship, eveLossData)

	// Check if the ship died in pochven
	if !isPochvenSystem(eveLossData.SolarSystemId) {
		if !userIsFc {
			return generateSimpleEmbed("❌ Outside Pochven ❌", "This ship was destroyed outside Pochven", "Link", link)
		} else {
			if shortenedWarning != "" {
				shortenedWarning += " | "
			}
			shortenedWarning += "Not Pochven"
			warning += "\tShip was not destroyed in Pochven\n"
		}
	}

	// Get the custom srp amount if relevant, only Fc's can pass in this value
	if customSrp != 0 {
		srp = customSrp
		if !userIsFc {
			return generateSimpleEmbed("❌ Permission Denied ❌", "Only an Fc can specify a custom Srp Amount", "Link", link)
		}
	}

	if srp == 0 {
		warning += "\tSrp value is zero! Please ask an FC to update this link with an srp amount.\n"
		if shortenedWarning != "" {
			shortenedWarning += " | "
		}
		shortenedWarning += "Zero Srp"
	}

	if warning != "" {
		warning = "Warning(s):\n" + warning + "Fc has overriden warnings"
	}

	//Submit the loss to the database, and report the result to the user
	loss = Losses{NickName: nickName, UserId: userID, Url: parsedLink, Srp: srp, ShipId: uint(eveLossData.ShipTypeId), ShipName: ship.Name, Warnings: shortenedWarning, KillMailId: loss.KillMailId}

	creationResult := db.Create(&loss)

	if creationResult.Error != nil {
		return generateSimpleEmbed("❌ Sql Error ❌", "The Zkill link you have submitted caused a Sql Error", "Link", link)
	} else {
		embed := &dg.MessageEmbed{

			Title:       "✅ Loss Submitted! ✅",
			Description: fmt.Sprintf("Your %s loss has been submitted", loss.ShipName),
			Fields: []*dg.MessageEmbedField{
				{
					Name:  "🔗 link",
					Value: parsedLink,
				},
				{

					Name:  "💰 Srp",
					Value: fmt.Sprintf("%d Million Isk", srp),
				},
			},
		}
		if warning != "" {
			embed.Fields = append(embed.Fields, &dg.MessageEmbedField{
				Name:  "Warnings",
				Value: warning,
			},
			)
		}
		return embed
	}
}

func regexMatchZkill(link string) string {
	re := regexp.MustCompile(`zkillboard\.com\/kill\/([0-9]+)`)
	results := re.FindStringSubmatch(link)
	if len(results) == 0 {
		return ""
	}
	return results[0]
}

func generateOptionMap(interaction *dg.InteractionCreate) *optionMap {
	options := interaction.ApplicationCommandData().Options
	optionMap := make(optionMap, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return &optionMap
}

func getLossIdFromLink(link string) string {
	re := regexp.MustCompile(`[0-9]+`)
	results := re.FindStringSubmatch(link)
	if len(results) == 0 {
		return ""
	}
	return results[0]
}

func getJsonFromZkill(link string) []ZkillLoss {
	lossId := getLossIdFromLink(link)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf(ZKILL_API_URL, lossId), nil)
	req.Header.Set("User-Agent", "Brave Pochven Srp Discord Bot") //Zkill requests that a "User-Agent" header be provided
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("ZKILL API GET Failed on link: %s", link)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ZKILL API GET Failed with status code %d on link: %s", resp.StatusCode, link)
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ZKILL API GET Failed on response parsing")
		return nil
	}

	loss := []ZkillLoss{}
	err = json.Unmarshal(data, &loss)
	if err != nil || len(loss) != 1 {
		log.Printf("Zkill JSON Unmarshal Error: %v", err)
		return nil
	}
	return loss
}

func getJsonFromEve(link string, killmailid uint64, hash string) EveLoss {
	resp, err := http.Get(fmt.Sprintf(EVE_API_URL, killmailid, hash))
	if err != nil {
		log.Printf("EVE API GET Failed on link: %s", link)
		return EveLoss{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("EVE API GET Failed with status code %d on link: %s", resp.StatusCode, link)
		return EveLoss{}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ZKILL API GET Failed on response parsing")
		return EveLoss{}
	}
	eveLoss := EveLoss{}
	err = json.Unmarshal(data, &eveLoss)
	if err != nil {
		log.Printf("Eve JSON Unmarshal Error: %v", err)
		return EveLoss{}
	}
	return eveLoss
}

// [{"killmail_id":115339013,"zkb":{"locationID":40004728,"hash":"d7b2b4bbb7e656c39528f55e28c313d26cdeab2b","fittedValue":66661524134.69,"droppedValue":1921508342.71,"destroyedValue":65209704820.74,"totalValue":67131213163.45,"points":1,"npc":false,"solo":false,"awox":false}}]
func getLossFromApi(link string) *Loss {
	zkillLoss := getJsonFromZkill(link)
	if zkillLoss == nil {
		return nil
	}

	eveLoss := getJsonFromEve(link, zkillLoss[0].KillmailId, zkillLoss[0].Data.Hash)

	if (eveLoss == EveLoss{}) {
		return nil
	}
	loss := Loss{
		KillmailId:    zkillLoss[0].KillmailId,
		Hash:          zkillLoss[0].Data.Hash,
		TotalValue:    zkillLoss[0].Data.TotalValue,
		SolarSystemId: eveLoss.SolarSystemId,
		ShipTypeId:    eveLoss.Victim.ShipTypeId,
	}
	return &loss
}

func getDoctrineShip(shipId uint) *DoctrineShips {
	ship := DoctrineShips{}
	db.Where("ship_id = ?", shipId).First(&ship)
	return &ship
}

func isUserFc(member *dg.User) bool {
	if member.ID == "416767410788630558" { //Jinx
		return true
	}
	res := db.Select("user_id").Where("user_id = ?", member.ID).First(&Administrators{})
	return res.Error == nil
}

func isUserSuperAdmin(member *dg.User) bool {
	if member.ID == "416767410788630558" { //Jinx
		return true
	}
	res := db.Select("user_id").Where("user_id = ? AND super_admin IS true", member.ID).First(&Administrators{})
	return res.Error == nil
}

func isPochvenSystem(systemId uint32) bool {
	for _, id := range PochvenSystems {
		if systemId == id {
			return true
		}
	}
	return false
}

func getShipNameFromId(shipId uint) string {
	doctrineShip := Ships{}
	db.Select("name").Where("ship_id = ?", shipId).First(&doctrineShip)
	if doctrineShip != (Ships{}) {
		return doctrineShip.Name
	}

	resp, err := http.Get(fmt.Sprintf(EVE_TYPE_URL, shipId))
	if err != nil {
		log.Printf("EVE API GET Failed on id: %v", shipId)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("EVE API GET Failed with status code %d on id: %d", resp.StatusCode, shipId)
		return ""
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ZKILL API GET Failed on response parsing")
		return ""
	}
	ship := Ship{}
	err = json.Unmarshal(data, &ship)
	if err != nil {
		log.Printf("Eve JSON Unmarshal Error: %v", err)
		return ""
	}
	db.Create(&Ships{Ship_ID: shipId, Name: ship.Name})
	return ship.Name
}

func getLossFromLink(link string) *Losses {
	loss := Losses{}
	db.Where("url = ?", link).Find(&loss)
	return &loss
}

func generateDoctrineShipEmbed(ships []DoctrineShips) *dg.MessageEmbed {
	embed := &dg.MessageEmbed{

		Title:       "✈️ Doctrine Ships 🛩️",
		Description: "All registered doctrine ships",
		Fields: []*dg.MessageEmbedField{
			{

				Name:   "✈️ Name",
				Inline: true,
			},
			{
				Name:   "🪪 Id",
				Inline: true,
			},
			{

				Name:   "💰 Srp",
				Inline: true,
			},
		},
	}

	for _, ship := range ships {
		embed.Fields[0].Value += fmt.Sprintf("%s\n", ship.Name)
		embed.Fields[1].Value += fmt.Sprintf("%d\n", ship.Ship_ID)
		embed.Fields[2].Value += fmt.Sprintf("%d\n", ship.Srp)
	}
	return embed
}

func generateSrpTotalEmbed(losses []Losses) *dg.MessageEmbed {

	type UserLossTotal struct {
		Total  uint64
		Losses []*Losses
	}
	embed := &dg.MessageEmbed{
		Title:       "💵 Srp Totals 💶",
		Description: "Loss totals and links are broken up per Capsuleer",
		Footer: &dg.MessageEmbedFooter{
			Text: "Any losses marked with an asterisk have warnings, and should be manually inspected.\n💰💰💰💰💰💰💰💰💰💰",
		},
	}

	lossesMap := make(map[string]*UserLossTotal)

	for _, loss := range losses {
		if opt, ok := lossesMap[loss.NickName]; ok {
			userLoss := opt
			userLoss.Losses = append(opt.Losses, &loss)
			userLoss.Total += loss.Srp
			lossesMap[loss.NickName] = userLoss
		} else {
			lossesMap[loss.NickName] = &UserLossTotal{Total: loss.Srp, Losses: []*Losses{&loss}}
		}
	}

	for nickName, userLoss := range lossesMap {
		userEmbed := &dg.MessageEmbedField{
			Name:   nickName,
			Inline: true,
		}
		srpEmbed := &dg.MessageEmbedField{
			Name:   "Srp",
			Inline: true,
		}
		for _, loss := range userLoss.Losses {
			srpEmbed.Value += fmt.Sprintf("%s: %d\n", loss.ShipName, loss.Srp)
			userEmbed.Value += loss.Url
			if loss.Warnings != "" {

				userEmbed.Value += "*"
			}
			userEmbed.Value += "\n"
			userLoss.Total += loss.Srp
		}
		embed.Fields = append(embed.Fields, userEmbed, srpEmbed)
		embed.Fields = append(embed.Fields, &dg.MessageEmbedField{Name: "Total Payout", Value: fmt.Sprintf("%d", userLoss.Total*1000000), Inline: true})
		embed.Fields = append(embed.Fields, &dg.MessageEmbedField{})
	}
	return embed

}

func generateSrpTotalEmbedUser(losses []Losses) *dg.MessageEmbed {
	nickName := losses[0].NickName

	srpTotal := uint64(0)
	embed := &dg.MessageEmbed{
		Title:       "💵 Srp Totals  💶",
		Description: fmt.Sprintf("Loss totals for Capsuleer: %s", nickName),
		Footer: &dg.MessageEmbedFooter{
			Text: "Any losses marked with an asterisk have warnings, and should be manually inspected.\n💰💰💰💰💰💰💰💰💰💰",
		},
	}

	//totalString := fmt.Sprintf("Losses|SRP for User: %s\n", nickName)

	userEmbed := &dg.MessageEmbedField{
		Name:   nickName,
		Inline: true,
	}
	srpEmbed := &dg.MessageEmbedField{
		Name:   "Srp",
		Inline: true,
	}

	for _, loss := range losses {
		srpEmbed.Value += fmt.Sprintf("%s: %d\n", loss.ShipName, loss.Srp)

		userEmbed.Value += loss.Url
		if loss.Warnings != "" {

			userEmbed.Value += "*"
		}
		userEmbed.Value += "\n"
		srpTotal += loss.Srp
	}
	embed.Fields = append(embed.Fields, userEmbed, srpEmbed)
	embed.Fields = append(embed.Fields, &dg.MessageEmbedField{Name: "Total Payout", Value: fmt.Sprintf("%d", srpTotal*1000000), Inline: true})

	return embed
}

func getDoctrineShipSrp(ship *DoctrineShips, eveLossData *Loss) uint64 {
	for _, id := range interdictorShipIds {
		if id == uint32(eveLossData.ShipTypeId) {
			roundedUp := math.Ceil((float64(eveLossData.TotalValue) / 1000000))
			return uint64(math.Min(float64(ship.Srp), roundedUp))
		}
	}
	return ship.Srp
}

func getNicknameFromUser(session *dg.Session, user *dg.User) string {
	if user.Bot {
		return user.Username
	}
	member, err := session.GuildMember(GUILD_ID, user.ID)
	if err != nil || member.Nick == "" {
		return user.Username
	}
	return member.Nick
}
