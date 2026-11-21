package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	staffOfTheMindMagusOracleID  = "1466687f-cfe1-4e6e-ad8e-dca5e8e4c06d"
	staffOfTheMindMagusText      = "Whenever you cast a blue spell or an Island you control enters, you gain 1 life."
	staffOfTheWildMagusOracleID  = "6bdff38d-0a52-42e5-86ea-edd7029b44f6"
	staffOfTheWildMagusText      = "Whenever you cast a green spell or a Forest you control enters, you gain 1 life."
	staffOfTheSunMagusOracleID   = "bd8bc927-e50a-4fab-9803-0602a841c7c2"
	staffOfTheSunMagusText       = "Whenever you cast a white spell or a Plains you control enters, you gain 1 life."
	staffOfTheDeathMagusOracleID = "c3fda16e-fd56-434e-91ac-561ae0482f43"
	staffOfTheDeathMagusText     = "Whenever you cast a black spell or a Swamp you control enters, you gain 1 life."
	staffOfTheFlameMagusOracleID = "fc71710a-8294-4370-89da-7ab1c2b38d19"
	staffOfTheFlameMagusText     = "Whenever you cast a red spell or a Mountain you control enters, you gain 1 life."
)

var StaffOfTheMindMagus = cardset.Entry{
	OracleID:     staffOfTheMindMagusOracleID,
	Name:         "Staff of the Mind Magus",
	New:          newStaffOfTheMindMagus,
	Verification: "mtg/rules/staff_of_the_magus_test.go: TestStaffOfTheMindMagusTriggersForBlueSpellsAndIslands",
}

var StaffOfTheWildMagus = cardset.Entry{
	OracleID:     staffOfTheWildMagusOracleID,
	Name:         "Staff of the Wild Magus",
	New:          newStaffOfTheWildMagus,
	Verification: "mtg/rules/staff_of_the_magus_test.go: TestStaffOfTheWildMagusTriggersForGreenSpellsAndForests",
}

var StaffOfTheSunMagus = cardset.Entry{
	OracleID:     staffOfTheSunMagusOracleID,
	Name:         "Staff of the Sun Magus",
	New:          newStaffOfTheSunMagus,
	Verification: "mtg/rules/staff_of_the_magus_test.go: TestStaffOfTheSunMagusTriggersForWhiteSpellsAndPlains",
}

var StaffOfTheDeathMagus = cardset.Entry{
	OracleID:     staffOfTheDeathMagusOracleID,
	Name:         "Staff of the Death Magus",
	New:          newStaffOfTheDeathMagus,
	Verification: "mtg/rules/staff_of_the_magus_test.go: TestStaffOfTheDeathMagusTriggersForBlackSpellsAndSwamps",
}

var StaffOfTheFlameMagus = cardset.Entry{
	OracleID:     staffOfTheFlameMagusOracleID,
	Name:         "Staff of the Flame Magus",
	New:          newStaffOfTheFlameMagus,
	Verification: "mtg/rules/staff_of_the_magus_test.go: TestStaffOfTheFlameMagusTriggersForRedSpellsAndMountains",
}

func newStaffOfTheMindMagus() *game.CardDef {
	return staffOfTheMagus("Staff of the Mind Magus", staffOfTheMindMagusText, color.Blue, types.Island)
}

func newStaffOfTheWildMagus() *game.CardDef {
	return staffOfTheMagus("Staff of the Wild Magus", staffOfTheWildMagusText, color.Green, types.Forest)
}

func newStaffOfTheSunMagus() *game.CardDef {
	return staffOfTheMagus("Staff of the Sun Magus", staffOfTheSunMagusText, color.White, types.Plains)
}

func newStaffOfTheDeathMagus() *game.CardDef {
	return staffOfTheMagus("Staff of the Death Magus", staffOfTheDeathMagusText, color.Black, types.Swamp)
}

func newStaffOfTheFlameMagus() *game.CardDef {
	return staffOfTheMagus("Staff of the Flame Magus", staffOfTheFlameMagusText, color.Red, types.Mountain)
}

func staffOfTheMagus(name, oracleText string, spellColor color.Color, landType types.Sub) *game.CardDef {
	gainLife := func() game.AbilityContent {
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{
				Player: game.ControllerReference(),
				Amount: game.Fixed(1),
			},
		}}}.Ability()
	}

	return &game.CardDef{
		ColorIdentity: color.NewIdentity(),
		CardFace: game.CardFace{
			Name:       name,
			ManaCost:   opt.Val(cost.Mana{cost.O(3)}),
			Types:      []types.Card{types.Artifact},
			OracleText: oracleText,
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: oracleText,
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ColorsAny: []color.Color{spellColor}},
						},
					},
					Content: gainLife(),
				},
				{
					Text: oracleText,
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{landType}},
						},
					},
					Content: gainLife(),
				},
			},
		},
	}
}
