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

type optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption

const ZKILL_API_URL string = "https://zkillboard.com/api/kills/killID/%v/"

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

type Loss struct {
	KillmailId uint64   `json:"killmail_id"`
	Data       LossData `json:"zkb"`
}

type LossData struct {
	LocationId     uint64  `json:"locationID"`
	Hash           string  `json:"hash"`
	FittedValue    float32 `json:"fittedValue"`
	DroppedValue   float32 `json:"droppedValue"`
	DestroyedValue float32 `json:"destroyedValue"`
	TotalValue     float32 `json:"totalValue"`
	Points         uint32  `json:"points"`
	Npc            bool    `json:"npc"`
	Solo           bool    `json:"solo"`
	AWox           bool    `json:"awox"`
}

// [{"killmail_id":115339013,"zkb":{"locationID":40004728,"hash":"d7b2b4bbb7e656c39528f55e28c313d26cdeab2b","fittedValue":66661524134.69,"droppedValue":1921508342.71,"destroyedValue":65209704820.74,"totalValue":67131213163.45,"points":1,"npc":false,"solo":false,"awox":false}}]
func getLossFromApi(link string) string {
	lossId := getLossIdFromLink(link)
	resp, err := http.Get(fmt.Sprintf(ZKILL_API_URL, lossId))
	if err != nil {
		log.Printf("API GET Failed on link: %s", link)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API GET Failed with status code %d on link: %s", resp.StatusCode, link)
		return ""
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("API GET Failed on response parsing")
		return ""
	}

	loss := []Loss{}
	err = json.Unmarshal(data, &loss)
	if err != nil || len(loss) != 1 {
		log.Printf("JSON Unmarshal Error: %v", err)
		return ""
	}

	log.Printf("Zkill Id: %v", loss[0].KillmailId)
	return string(data)
}
