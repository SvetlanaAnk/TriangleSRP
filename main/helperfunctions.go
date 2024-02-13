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

func sendInteractionResponse(session *dg.Session, interaction *dg.InteractionCreate, message string) {
	session.InteractionRespond(interaction.Interaction, &dg.InteractionResponse{
		Type: dg.InteractionResponseChannelMessageWithSource,
		Data: &dg.InteractionResponseData{
			Content: message,
		},
	})
}

func addKill(userName string, link string, userIsFc bool, customSrp uint64) string {
	warning := ""
	shortenedWarning := ""
	srp := uint64(1)

	// Verify that the link is valid, and pass it through ToLower() and regex
	parsedLink := regexMatchZkill(strings.ToLower(link))

	if parsedLink == "" {
		return fmt.Sprintf("Invalid Zkill format: %v", link)
	}

	// Check if this loss already exists on the Srp sheet
	loss := *getLossFromLink(parsedLink)
	if loss != (Losses{}) {
		return fmt.Sprintf("Link has already been submitted\n%v", link)
	}

	//Query the Zkill and Eve api's for needed information
	eveLossData := getLossFromApi(parsedLink)

	ship := getDoctrineShip(uint(eveLossData.ShipTypeId))

	// Check if the ship is a doctrine ship.
	if *ship == (DoctrineShips{}) {
		if !userIsFc {
			return fmt.Sprintf("%v: Ship is not a valid doctrine ship, please ask an FC to override", link)
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
			return fmt.Sprintf("%s: This ship was destroyed outside of Pochven, please ask an FC to override", link)
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
			return fmt.Sprintf("%s: Only an FC can specify a custom Srp amount.", link)
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
	loss = Losses{UserName: userName, Url: parsedLink, Srp: srp, ShipId: uint(eveLossData.ShipTypeId), ShipName: ship.Name, Warnings: shortenedWarning}

	creationResult := db.Create(&loss)

	if creationResult.Error != nil {
		return fmt.Sprintf("SQL Error submitting Link. %v", link)
	} else {
		return fmt.Sprintf("Submitted successfully\nLoss: %s\nAmount: %v million isk\nCapsuleer: %v\n%s", link, srp, userName, warning)
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

func isUserFc(member *dg.Member) bool {
	for _, role := range member.Roles {
		if role == "FC" {
			return true
		}
	}
	return member.User.Username == "theblob8584"
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
	doctrineShip := *getDoctrineShip(shipId)
	if doctrineShip != (DoctrineShips{}) {
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
	return ship.Name
}

func getLossFromLink(link string) *Losses {
	loss := Losses{}
	db.Where("url = ?", link).Find(&loss)
	return &loss
}

func generateDoctrineShipString(ships []DoctrineShips) string {
	shipString := ""
	for _, ship := range ships {
		shipString += fmt.Sprintf("Name: %s ID: %d Srp: %d Million isk\n", ship.Name, ship.Ship_ID, ship.Srp)
	}
	return shipString
}

func generateSrpTotalString(losses []Losses, printZkill bool, printWarnings bool) string {

	type UserLossTotal struct {
		Total  uint64
		Losses []Losses
	}

	totalsString := "Loss Totals:\n"

	lossesMap := make(map[string]UserLossTotal)

	for _, loss := range losses {
		var userLoss UserLossTotal
		if val, ok := lossesMap[loss.UserName]; ok {
			userLoss = val
		} else {
			userLoss = UserLossTotal{}
		}
		userLoss.Total += loss.Srp
		userLoss.Losses = append(userLoss.Losses, loss)

		lossesMap[loss.UserName] = userLoss
	}

	for userName, userLoss := range lossesMap {
		if !printZkill {
			continue
		}
		totalsString += fmt.Sprintf("User: %s\nLosses:\n", userName)
		for _, loss := range userLoss.Losses {
			totalsString += fmt.Sprintf("\t%s\n", loss.Url)
			if loss.Warnings != "" && printWarnings {
				totalsString += fmt.Sprintf("\t\tWarnings: %s\n", loss.Warnings)
			}
		}
		totalsString += fmt.Sprintf("Total: %d isk\n\n", userLoss.Total*1000000)
	}

	return totalsString
}

func generateSrpTotalForUser(losses []Losses) string {
	srpTotal := uint64(0)
	totalString := fmt.Sprintf("Losses|SRP for User: %s\n", losses[0].UserName)

	for _, loss := range losses {
		totalString += fmt.Sprintf("\tShip: %s Srp: %d Million Isk\n\t\tZkill: %s\n", loss.ShipName, loss.Srp, loss.Url)
		if loss.Warnings != "" {
			totalString += fmt.Sprintf("\t\t%s", loss.Warnings)
		}

		srpTotal += loss.Srp
	}

	totalString += fmt.Sprintf("Total Srp: %d", srpTotal*1000000)

	return totalString
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
