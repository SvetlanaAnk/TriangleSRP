package main

import dg "github.com/bwmarrin/discordgo"

type optionMap map[string]*dg.ApplicationCommandInteractionDataOption

const ZKILL_API_URL string = "https://zkillboard.com/api/kills/killID/%v/"
const EVE_API_URL string = "https://esi.evetech.net/latest/killmails/%d/%s/?datasource=tranquility"
const EVE_TYPE_URL string = "https://esi.evetech.net/latest/universe/types/%d/?datasource=tranquility&language=en"
const DATABASE_FILE string = "srpmain.sqlite"

var SRP_CHANNEL_MAP = make(map[string]string)

type Loss struct {
	KillmailId    uint64
	LocationId    uint64
	Hash          string
	TotalValue    float32
	SolarSystemId uint32
	ShipTypeId    uint32
}
type ZkillLoss struct {
	KillmailId uint64        `json:"killmail_id"`
	Data       ZkillLossData `json:"zkb"`
}

type ZkillLossData struct {
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

type Ship struct {
	Name string `json:"name"`
}

var PochvenSystems []uint32 = []uint32{
	30000157,
	30000192,
	30001372,
	30001445,
	30002079,
	30002737,
	30005005,
	30010141,
	30031392,
	30000021,
	30001413,
	30002225,
	30002411,
	30002770,
	30003495,
	30003504,
	30040141,
	30045328,
	30000206,
	30001381,
	30002652,
	30002702,
	30002797,
	30003046,
	30005029,
	30020141,
	30045329,
}

var interdictorShipIds []uint32 = []uint32{
	22464, // Flycatcher
	22456, // Sabre
	22460, // Eris
	22452, // Heretic
	37482, // Stork
	37481, // Pontifex
	37483, // Magus
	37480, // Bifrost
}
