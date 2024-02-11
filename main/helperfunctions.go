package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

func sendInteractionResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate, message string) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func regexMatchZkill(link string) string {
	re := regexp.MustCompile(`zkillboard\.com\/kill\/([0-9]+)`)
	results := re.FindStringSubmatch(link)
	if len(results) == 0 {
		return ""
	}
	return results[0]
}

func generateOptionMap(interaction *discordgo.InteractionCreate) *optionMap {
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

func getJsonFromZkill(link string) []Loss {
	lossId := getLossIdFromLink(link)
	resp, err := http.Get(fmt.Sprintf(ZKILL_API_URL, lossId))
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

	loss := []Loss{}
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
func getLossFromApi(link string) ([]Loss, EveLoss) {
	loss := getJsonFromZkill(link)
	if loss == nil {
		return nil, EveLoss{}
	}

	eveLoss := getJsonFromEve(link, loss[0].KillmailId, loss[0].Data.Hash)

	if (eveLoss == EveLoss{}) {
		return nil, EveLoss{}
	}

	return loss, eveLoss
}

func getDoctrineShip(shipId uint) *DoctrineShips {
	ship := DoctrineShips{}
	db.Where("id = ?", shipId).First(&ship)
	return &ship
}

func isUserFc(interaction *discordgo.InteractionCreate) bool {
	for _, role := range interaction.Member.Roles {
		if role == "FC" {
			return true
		}
	}
	return interaction.Member.User.Username == "theblob8584"
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
