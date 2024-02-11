package main

import (
	"github.com/bwmarrin/discordgo"
)

var (
	integerOptionMinValue = 1.0
	commands              = []*discordgo.ApplicationCommand{
		{
			Name:        "add-kill",
			Description: "Add a kill to the SRP sheet",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The Zkill link of the lost ship",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user who lost the ship, or should receive the srp",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "srp",
					Description: "Set a SRP amount for this kill, in millions of isk",
					MinValue:    &integerOptionMinValue,
					Required:    false,
				},
			},
		},
	}
)
