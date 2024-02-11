package main

import (
	"github.com/bwmarrin/discordgo"
)

var (
	integerOptionMinValue = 1.0
	commands              = []*discordgo.ApplicationCommand{
		{
			Name:        "add-loss",
			Description: "Add a loss to the Srp sheet",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The Zkill link of the loss",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user who lost the ship, or should receive Srp",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "srp",
					Description: "Set an Srp amount for this kill, in millions of isk",
					MinValue:    &integerOptionMinValue,
					Required:    false,
				},
			},
		},
		{
			Name:        "set-ship-srp",
			Description: "Add or update a doctrine ship with an Srp value",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "ship-id",
					Description: "The ship id. Check the ship's zkill page url for this id",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "srp",
					Description: "Max srp payout, in millions of isk",
					MinValue:    &integerOptionMinValue,
					Required:    true,
				},
			},
		},
		{
			Name:        "remove-loss",
			Description: "Remove a loss from the Srp sheet",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The zkill link of the loss",
					Required:    true,
				},
			},
		},
		{
			Name:        "update-loss",
			Description: "Update the Srp on a loss",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The zkill link of the lost ship",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user who lost the ship, or should receive the Srp",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "srp",
					Description: "New Srp Payout, in millions of isk",
					MinValue:    &integerOptionMinValue,
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "paid",
					Description: "New paid status: true or false",
					Required:    false,
				},
			},
		},
		{
			Name:        "paid",
			Description: "Mark a loss as paid",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The zkill link of the lost ship",
					Required:    true,
				},
			},
		},
	}
)
