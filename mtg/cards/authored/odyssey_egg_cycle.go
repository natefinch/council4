package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	mossfireEggOracleID    = "364b3231-c0e7-45a8-90b9-a1cdb584dd7c"
	mossfireEggText        = "{2}, {T}, Sacrifice this artifact: Add {R}{G}. Draw a card. (Activate only as an instant.)"
	sungrassEggOracleID    = "80a49a1b-a202-4c14-b093-dc76eb0f42c7"
	sungrassEggText        = "{2}, {T}, Sacrifice this artifact: Add {G}{W}. Draw a card. (Activate only as an instant.)"
	shadowbloodEggOracleID = "9b58e7fa-4259-4d6b-8b5d-33fb37fe489f"
	shadowbloodEggText     = "{2}, {T}, Sacrifice this artifact: Add {B}{R}. Draw a card. (Activate only as an instant.)"
	darkwaterEggOracleID   = "c4893a24-3cbc-4011-bef4-50e0c4dce16e"
	darkwaterEggText       = "{2}, {T}, Sacrifice this artifact: Add {U}{B}. Draw a card. (Activate only as an instant.)"
)

var (
	MossfireEgg = cardset.Entry{
		OracleID: mossfireEggOracleID,
		Name:     "Mossfire Egg",
		New: func() *game.CardDef {
			return newOdysseyEgg("Mossfire Egg", mossfireEggText, color.NewIdentity(color.Red, color.Green), mana.R, mana.G)
		},
		Verification: "mtg/rules/odyssey_egg_cycle_test.go: TestMossfireEggPaysCostsThenAddsManaAndDraws; " +
			"Odyssey printing 1fd66433-b0e9-453d-9a27-ea0efb158dac; Scryfall rulings retrieved 2026-11-21 (one WotC ruling dated 2008-08-01); " +
			"CR 2026-09-25 and 2026-11-06: 113.3b, 118.1-3, 121.1, 601.2f-h, 602.1a-b, 602.2a-b, 602.5e, 605.1a, 605.3b, 608.2c",
	}
	SungrassEgg = cardset.Entry{
		OracleID: sungrassEggOracleID,
		Name:     "Sungrass Egg",
		New: func() *game.CardDef {
			return newOdysseyEgg("Sungrass Egg", sungrassEggText, color.NewIdentity(color.Green, color.White), mana.G, mana.W)
		},
		Verification: "mtg/rules/odyssey_egg_cycle_test.go: TestSungrassEggPaysCostsThenAddsManaAndDraws; " +
			"Odyssey printing 70e888f9-c583-478a-9ef0-2e89db8a8dbb; Scryfall rulings retrieved 2026-11-21 (one WotC ruling dated 2008-08-01); " +
			"CR 2026-09-25 and 2026-11-06: 113.3b, 118.1-3, 121.1, 601.2f-h, 602.1a-b, 602.2a-b, 602.5e, 605.1a, 605.3b, 608.2c",
	}
	ShadowbloodEgg = cardset.Entry{
		OracleID: shadowbloodEggOracleID,
		Name:     "Shadowblood Egg",
		New: func() *game.CardDef {
			return newOdysseyEgg("Shadowblood Egg", shadowbloodEggText, color.NewIdentity(color.Black, color.Red), mana.B, mana.R)
		},
		Verification: "mtg/rules/odyssey_egg_cycle_test.go: TestShadowbloodEggPaysCostsThenAddsManaAndDraws; " +
			"Odyssey printing 044bbf15-e71d-4f47-bac3-bba8021118bd; Scryfall rulings retrieved 2026-11-21 (one WotC ruling dated 2008-08-01); " +
			"CR 2026-09-25 and 2026-11-06: 113.3b, 118.1-3, 121.1, 601.2f-h, 602.1a-b, 602.2a-b, 602.5e, 605.1a, 605.3b, 608.2c",
	}
	DarkwaterEgg = cardset.Entry{
		OracleID: darkwaterEggOracleID,
		Name:     "Darkwater Egg",
		New: func() *game.CardDef {
			return newOdysseyEgg("Darkwater Egg", darkwaterEggText, color.NewIdentity(color.Blue, color.Black), mana.U, mana.B)
		},
		Verification: "mtg/rules/odyssey_egg_cycle_test.go: TestDarkwaterEggPaysCostsThenAddsManaAndDraws; " +
			"Odyssey printing fb988572-1e1d-4b1d-8dc7-78bda966554e; Scryfall rulings retrieved 2026-11-21 (one WotC ruling dated 2008-08-01); " +
			"CR 2026-09-25 and 2026-11-06: 113.3b, 118.1-3, 121.1, 601.2f-h, 602.1a-b, 602.2a-b, 602.5e, 605.1a, 605.3b, 608.2c",
	}
)

func newOdysseyEgg(name, text string, identity color.Identity, first, second mana.Color) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: identity,
		CardFace: game.CardFace{
			Name:       name,
			ManaCost:   opt.Val(cost.Mana{cost.O(1)}),
			Types:      []types.Card{types.Artifact},
			OracleText: text,
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:     text,
				ManaCost: opt.Val(cost.Mana{cost.O(2)}),
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalTap},
					{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this artifact", Amount: 1},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: first}},
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: second}},
					{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
				}}.Ability(),
			}},
		},
	}
}
