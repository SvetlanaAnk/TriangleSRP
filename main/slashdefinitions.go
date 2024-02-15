package main

import (
	dg "github.com/bwmarrin/discordgo"
)

var (
	integerOptionMinValue = 1.0
	commands              = []*dg.ApplicationCommand{
		{
			Name:        "add-loss",
			Description: "Add a loss to the Srp sheet",
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The Zkill link of the loss",
					Required:    true,
				},
				{
					Type:        dg.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user who lost the ship, or should receive Srp",
					Required:    false,
				},
				{
					Type:        dg.ApplicationCommandOptionInteger,
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
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionInteger,
					Name:        "ship-id",
					Description: "The ship id. Check the ship's zkill page url for this id",
					Required:    true,
				},
				{
					Type:        dg.ApplicationCommandOptionInteger,
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
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The zkill link of the loss",
					Required:    true,
				},
			},
		},
		{
			Name:        "update-loss",
			Description: "Update the Srp on a loss",
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The zkill link of the lost ship",
					Required:    true,
				},
				{
					Type:        dg.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user who lost the ship, or should receive the Srp",
					Required:    false,
				},
				{
					Type:        dg.ApplicationCommandOptionInteger,
					Name:        "srp",
					Description: "New Srp Payout, in millions of isk",
					MinValue:    &integerOptionMinValue,
					Required:    false,
				},
				{
					Type:        dg.ApplicationCommandOptionBoolean,
					Name:        "paid",
					Description: "New paid status: true or false",
					Required:    false,
				},
			},
		},
		{
			Name:        "paid",
			Description: "Mark a loss as paid",
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionString,
					Name:        "link",
					Description: "The zkill link of the lost ship",
					Required:    true,
				},
			},
		},
		{
			Name:        "mark-user-paid",
			Description: "Mark a user's losses as paid",
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user whose losses will be marked as paid",
					Required:    false,
				},
			},
		},
		{
			Name:        "print-ships",
			Description: "Print all current doctrine ships and their srp amounts.",
		},
		{
			Name:        "srp-totals",
			Description: "Print srp totals ",
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionBoolean,
					Name:        "include-zkill",
					Description: "Should I include zkill links?",
					Required:    false,
				},
				{
					Type:        dg.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Fetch the srp totals for a specific user",
					Required:    false,
				},
				{
					Type:        dg.ApplicationCommandOptionUser,
					Name:        "include-warnings",
					Description: "Should I include loss warnings?",
					Required:    false,
				},
			},
		},
		{
			Name:        "user-srp-total",
			Description: "Get the losses and srp total for a user",
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Fetch totals for a specific player",
					Required:    false,
				},
			},
		},
		{
			Name:        "remove-ship",
			Description: "Remove a doctrine ship",
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionInteger,
					Name:        "ship-id",
					Description: "The ship id. Check the ship's zkill page url for this id",
					Required:    true,
				},
			},
		},
		{
			Name:        "set-channel",
			Description: "Set the Srp channel the bot should listen to",
		},
		{
			Name:        "add-fc",
			Description: "Add a fleet commander",
			Options: []*dg.ApplicationCommandOption{
				{
					Type:        dg.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user who will become recognized as an fc",
					Required:    true,
				},
			},
		},
		{
			Name:        "srp-paid",
			Description: "Mark all losses as paid",
		},
	}
)
