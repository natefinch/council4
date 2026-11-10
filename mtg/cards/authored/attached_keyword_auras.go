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
	fearAuraOracleID      = "355bbe9b-59bf-470f-8600-410af4c7fe18"
	lifelinkAuraOracleID  = "4fe8316a-9cff-43e9-a1d5-993a0c9daf3a"
	vigilanceAuraOracleID = "70570170-be76-4c56-9151-c4b6e253f462"

	fearAuraText      = "Enchant creature (Target a creature as you cast this. This card enters attached to that creature.)\nEnchanted creature has fear. (It can't be blocked except by artifact creatures and/or black creatures.)"
	lifelinkAuraText  = "Enchant creature\nEnchanted creature has lifelink. (Damage dealt by the creature also causes its controller to gain that much life.)"
	vigilanceAuraText = "Enchant creature\nEnchanted creature has vigilance. (Attacking doesn't cause it to tap.)"
)

var (
	FearAura = cardset.Entry{
		OracleID: fearAuraOracleID,
		Name:     "Fear",
		New:      newFearAura,
		Verification: "mtg/rules/attached_keyword_auras_test.go: " +
			"TestFearAuraGrantsFearInCombat",
	}

	LifelinkAura = cardset.Entry{
		OracleID: lifelinkAuraOracleID,
		Name:     "Lifelink",
		New:      newLifelinkAura,
		Verification: "mtg/rules/attached_keyword_auras_test.go: " +
			"TestLifelinkAuraGrantsLifeToEnchantedCreatureController",
	}

	VigilanceAura = cardset.Entry{
		OracleID: vigilanceAuraOracleID,
		Name:     "Vigilance",
		New:      newVigilanceAura,
		Verification: "mtg/rules/attached_keyword_auras_test.go: " +
			"TestVigilanceAuraAllowsEnchantedCreatureToAttackUntapped",
	}
)

func newFearAura() *game.CardDef {
	return newAttachedKeywordAura(
		"Fear", cost.Mana{cost.B, cost.B}, color.Black, fearAuraText, "Enchanted creature has fear.", game.Fear,
	)
}

func newLifelinkAura() *game.CardDef {
	return newAttachedKeywordAura(
		"Lifelink", cost.Mana{cost.W}, color.White, lifelinkAuraText, "Enchanted creature has lifelink.", game.Lifelink,
	)
}

func newVigilanceAura() *game.CardDef {
	return newAttachedKeywordAura(
		"Vigilance", cost.Mana{cost.W}, color.White, vigilanceAuraText, "Enchanted creature has vigilance.", game.Vigilance,
	)
}

func newAttachedKeywordAura(name string, manaCost cost.Mana, cardColor color.Color, oracleText, abilityText string, keyword game.Keyword) *game.CardDef {
	attached := game.AttachedObjectGroup(game.SourcePermanentReference())
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		ManaCost:   opt.Val(manaCost),
		Types:      []types.Card{types.Enchantment},
		Subtypes:   []types.Sub{types.Aura},
		Colors:     []color.Color{cardColor},
		OracleText: oracleText,
		StaticAbilities: []game.StaticAbility{
			{
				Text: "Enchant creature",
				KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{
					Target: game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					},
				}},
			},
			{
				Text: abilityText,
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					Group:       attached,
					AddKeywords: []game.Keyword{keyword},
				}},
			},
		},
	}}
}
