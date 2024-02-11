package main

import "github.com/bwmarrin/discordgo"

type optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption

const ZKILL_API_URL string = "https://zkillboard.com/api/kills/killID/%v/"
const EVE_API_URL string = "https://esi.evetech.net/latest/killmails/%d/%s/?datasource=tranquility"
const DATABASE_FILE string = "srpmain.sqlite"

type Loss struct {
	KillmailId uint64   `json:"killmail_id"`
	Data       LossData `json:"zkb"`
}

type LossData struct {
	LocationId uint64  `json:"locationID"`
	Hash       string  `json:"hash"`
	TotalValue float32 `json:"totalValue"`
}

type EveLoss struct {
	SolarSystemId uint32 `json:"solar_system_id"`
	Victim        Victim `json:"victim"`
}

type Victim struct {
	ShipTypeId uint32 `json:"ship_type_id"`
}
