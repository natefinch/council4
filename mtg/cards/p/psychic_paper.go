package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PsychicPaper is the card definition for Psychic Paper.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	As this Equipment becomes attached to a creature, choose a creature card name and a creature type.
//	Equipped creature has ward {1}, it can't be blocked, and its name and creature type are the last chosen name and creature type.
//	Equip {2}
var PsychicPaper = newPsychicPaper

func newPsychicPaper() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Psychic Paper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has ward {1}, it can't be blocked, and its name and creature type are the last chosen name and creature type.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(1)})),
							},
						},
						game.ContinuousEffect{
							Duration:                game.DurationPermanent,
							Group:                   game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:                   game.LayerText,
							SetNameFromSourceChoice: game.AttachmentCardNameChoiceKey,
						},
						game.ContinuousEffect{
							Duration:                   game.DurationPermanent,
							Group:                      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:                      game.LayerType,
							SetSubtypeFromSourceChoice: game.AttachmentSubtypeChoiceKey,
							SetSubtypeChoiceType:       types.Creature,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBeBlocked,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.AttachmentChoicesReplacement("As this Equipment becomes attached to a creature, choose a creature card name and a creature type.", types.Creature, types.Creature),
			},
			OracleText: `
			As this Equipment becomes attached to a creature, choose a creature card name and a creature type.
			Equipped creature has ward {1}, it can't be blocked, and its name and creature type are the last chosen name and creature type.
			Equip {2}
		`,
		},
	}
}
