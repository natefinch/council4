package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// This file defines the token definitions dungeon rooms create. They live in
// mtg/game so the dungeon graph definitions can reference them without importing
// mtg/cards (which would form an import cycle, since token defs there import
// mtg/game). Each matches the printed token a room creates exactly.

// dungeonGoblinToken is the "1/1 red Goblin creature token" created by Lost Mine
// of Phandelver's Goblin Lair.
var dungeonGoblinToken = &CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: CardFace{
		Name:      "Goblin",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin},
		Power:     opt.Val(PT{Value: 1}),
		Toughness: opt.Val(PT{Value: 1}),
	},
}

// dungeonTreasureToken is the "Treasure token" created by several rooms. It is a
// Treasure artifact with the standard "{T}, Sacrifice this artifact: Add one
// mana of any color." mana ability.
var dungeonTreasureToken = &CardDef{
	CardFace: CardFace{
		Name:     "Treasure",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Treasure},
		ManaAbilities: []ManaAbility{
			{
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalTap},
					{
						Kind:               cost.AdditionalSacrificeSource,
						Text:               "Sacrifice this artifact",
						Amount:             1,
						MatchPermanentType: true,
						PermanentType:      types.Artifact,
					},
				},
				Content: Mode{
					Sequence: []Instruction{
						{
							Primitive: Choose{
								Choice: ResolutionChoice{
									Kind:   ResolutionChoiceMana,
									Prompt: "Choose a color",
									Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
								PublishChoice: ChoiceKey("oracle-mana-color"),
							},
						},
						{
							Primitive: AddMana{
								Amount:     Fixed(1),
								ChoiceFrom: ChoiceKey("oracle-mana-color"),
							},
						},
					},
				}.Ability(),
			},
		},
	},
}

// dungeonSkeletonToken is the "1/1 black Skeleton creature token" created by
// Dungeon of the Mad Mage's Muiral's Graveyard.
var dungeonSkeletonToken = &CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: CardFace{
		Name:      "Skeleton",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Skeleton},
		Power:     opt.Val(PT{Value: 1}),
		Toughness: opt.Val(PT{Value: 1}),
	},
}

// dungeonMenaceSkeletonToken is the "4/1 black Skeleton creature token with
// menace" created by Undercity's Catacombs.
var dungeonMenaceSkeletonToken = &CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: CardFace{
		Name:            "Skeleton",
		Colors:          []color.Color{color.Black},
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Skeleton},
		Power:           opt.Val(PT{Value: 4}),
		Toughness:       opt.Val(PT{Value: 1}),
		StaticAbilities: []StaticAbility{MenaceStaticBody},
	},
}

// dungeonAtropalToken is "The Atropal, a legendary 4/4 black God Horror creature
// token with deathtouch" created by Tomb of Annihilation's Cradle of the Death
// God.
var dungeonAtropalToken = &CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: CardFace{
		Name:            "The Atropal",
		Colors:          []color.Color{color.Black},
		Supertypes:      []types.Super{types.Legendary},
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.God, types.Horror},
		Power:           opt.Val(PT{Value: 4}),
		Toughness:       opt.Val(PT{Value: 4}),
		StaticAbilities: []StaticAbility{DeathtouchStaticBody},
	},
}

// dungeonKnightToken is the "2/2 white Knight creature token" created by Baldur's
// Gate Wilderness's Emerald Grove.
var dungeonKnightToken = &CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: CardFace{
		Name:      "Knight",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Knight},
		Power:     opt.Val(PT{Value: 2}),
		Toughness: opt.Val(PT{Value: 2}),
	},
}

// dungeonFaerieDragonToken is the "1/1 blue Faerie Dragon creature token with
// flying" created by Baldur's Gate Wilderness's Ebonlake Grotto.
var dungeonFaerieDragonToken = &CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: CardFace{
		Name:            "Faerie Dragon",
		Colors:          []color.Color{color.Blue},
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Faerie, types.Dragon},
		Power:           opt.Val(PT{Value: 1}),
		Toughness:       opt.Val(PT{Value: 1}),
		StaticAbilities: []StaticAbility{FlyingStaticBody},
	},
}
