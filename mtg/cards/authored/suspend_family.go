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
	ancestralVisionOracleID    = "9728dec9-d482-4c7a-8cdc-44d010dc878d"
	searchForTomorrowOracleID  = "9cdc9f99-c6fa-40cd-90b0-e47d43a8cc3c"
	visceridDeepwalkerOracleID = "9dea848b-d758-46d0-829d-b096061fd460"
	keldonHalberdierOracleID   = "a1472a2d-e6f7-4d65-8bb2-4ecfae2d2764"
	corpulentCorpseOracleID    = "b3a84b75-a6a2-4521-a9f1-92596796c703"
	errantEphemeronOracleID    = "c77b89ec-4514-4e62-a508-2012f38a3e15"
	moxTantaliteOracleID       = "cc319819-68f9-4c1e-8d60-25a5e3e50669"
	infiltratorIlKorOracleID   = "dd491ac7-3c6c-4a47-94af-071d1edc5eb2"
	ithHighArcanistOracleID    = "eb933723-b879-44c8-bbc8-d8c84ce5ab12"
	riftSowerOracleID          = "f9deeea1-bf80-4832-aac3-373fc1f4b41d"

	ancestralVisionOracleText     = "Suspend 4—{U} (Rather than cast this card from your hand, pay {U} and exile it with four time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost.)\nTarget player draws three cards."
	searchForTomorrowOracleText   = "Search your library for a basic land card, put it onto the battlefield, then shuffle.\nSuspend 2—{G} (Rather than cast this card from your hand, you may pay {G} and exile it with two time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost.)"
	visceridDeepwalkerAbilityText = "{U}: This creature gets +1/+0 until end of turn."
	visceridDeepwalkerOracleText  = "{U}: This creature gets +1/+0 until end of turn.\nSuspend 4—{U} (Rather than cast this card from your hand, you may pay {U} and exile it with four time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost. It has haste.)"
	keldonHalberdierOracleText    = "First strike\nSuspend 4—{R} (Rather than cast this card from your hand, you may pay {R} and exile it with four time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost. It has haste.)"
	corpulentCorpseOracleText     = "Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)\nSuspend 5—{B} (Rather than cast this card from your hand, you may pay {B} and exile it with five time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost. It has haste.)"
	errantEphemeronOracleText     = "Flying\nSuspend 4—{1}{U} (Rather than cast this card from your hand, you may pay {1}{U} and exile it with four time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost. It has haste.)"
	moxTantaliteManaText          = "{T}: Add one mana of any color."
	moxTantaliteOracleText        = "Suspend 3—{0} (Rather than cast this card from your hand, pay {0} and exile it with three time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost.)\n{T}: Add one mana of any color."
	infiltratorIlKorOracleText    = "Shadow (This creature can block or be blocked by only creatures with shadow.)\nSuspend 2—{1}{U} (Rather than cast this card from your hand, you may pay {1}{U} and exile it with two time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost. It has haste.)"
	ithHighArcanistAbilityText    = "{T}: Untap target attacking creature. Prevent all combat damage that would be dealt to and dealt by that creature this turn."
	ithHighArcanistOracleText     = "Vigilance\n" + ithHighArcanistAbilityText + "\nSuspend 4—{W}{U}"
	riftSowerManaText             = "{T}: Add one mana of any color."
	riftSowerOracleText           = riftSowerManaText + "\nSuspend 2—{G} (Rather than cast this card from your hand, you may pay {G} and exile it with two time counters on it. At the beginning of your upkeep, remove a time counter. When the last is removed, you may cast it without paying its mana cost. It has haste.)"
)

var (
	AncestralVision = cardset.Entry{
		OracleID:     ancestralVisionOracleID,
		Name:         "Ancestral Vision",
		New:          newAncestralVision,
		Verification: "mtg/rules/suspend_family_test.go: TestAncestralVisionSuspendTargetsPlayerAndDrawsThree",
	}
	SearchForTomorrow = cardset.Entry{
		OracleID:     searchForTomorrowOracleID,
		Name:         "Search for Tomorrow",
		New:          newSearchForTomorrow,
		Verification: "mtg/rules/suspend_family_test.go: TestSearchForTomorrowSuspendFindsBasicLand",
	}
	VisceridDeepwalker = cardset.Entry{
		OracleID:     visceridDeepwalkerOracleID,
		Name:         "Viscerid Deepwalker",
		New:          newVisceridDeepwalker,
		Verification: "mtg/rules/suspend_family_test.go: TestVisceridDeepwalkerSuspendPump",
	}
	KeldonHalberdier = cardset.Entry{
		OracleID:     keldonHalberdierOracleID,
		Name:         "Keldon Halberdier",
		New:          newKeldonHalberdier,
		Verification: "mtg/rules/suspend_family_test.go: TestKeldonHalberdierSuspendFirstStrike",
	}
	CorpulentCorpse = cardset.Entry{
		OracleID:     corpulentCorpseOracleID,
		Name:         "Corpulent Corpse",
		New:          newCorpulentCorpse,
		Verification: "mtg/rules/suspend_family_test.go: TestCorpulentCorpseSuspendFear",
	}
	ErrantEphemeron = cardset.Entry{
		OracleID:     errantEphemeronOracleID,
		Name:         "Errant Ephemeron",
		New:          newErrantEphemeron,
		Verification: "mtg/rules/suspend_family_test.go: TestErrantEphemeronSuspendFlying",
	}
	MoxTantalite = cardset.Entry{
		OracleID:     moxTantaliteOracleID,
		Name:         "Mox Tantalite",
		New:          newMoxTantalite,
		Verification: "mtg/rules/suspend_family_test.go: TestMoxTantaliteZeroCostSuspendAndAnyColorMana",
	}
	InfiltratorIlKor = cardset.Entry{
		OracleID:     infiltratorIlKorOracleID,
		Name:         "Infiltrator il-Kor",
		New:          newInfiltratorIlKor,
		Verification: "mtg/rules/suspend_family_test.go: TestInfiltratorIlKorSuspendShadow",
	}
	IthHighArcanist = cardset.Entry{
		OracleID:     ithHighArcanistOracleID,
		Name:         "Ith, High Arcanist",
		New:          newIthHighArcanist,
		Verification: "mtg/rules/suspend_family_test.go: TestIthHighArcanistVigilanceUntapsAndPreventsCombatDamage",
	}
	RiftSower = cardset.Entry{
		OracleID:     riftSowerOracleID,
		Name:         "Rift Sower",
		New:          newRiftSower,
		Verification: "mtg/rules/suspend_family_test.go: TestRiftSowerSuspendAnyColorMana",
	}
)

func suspendFamilyAbility(manaCost cost.Mana, counters int) game.StaticAbility {
	return game.StaticAbility{KeywordAbilities: []game.KeywordAbility{
		game.SuspendKeyword{Cost: manaCost, TimeCounters: counters},
	}}
}

func newAncestralVision() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Ancestral Vision",
		Types:      []types.Card{types.Sorcery},
		Colors:     []color.Color{color.Blue},
		OracleText: ancestralVisionOracleText,
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target player",
				Allow:      game.TargetAllowPlayer,
			}},
			Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(3), Player: game.TargetPlayerReference(0)},
			}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{
			suspendFamilyAbility(cost.Mana{cost.U}, 4),
		},
	}}
}

func newSearchForTomorrow() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Search for Tomorrow",
		ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.G}),
		Types:      []types.Card{types.Sorcery},
		Colors:     []color.Color{color.Green},
		OracleText: searchForTomorrowOracleText,
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Search{
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone:  zone.Library,
					Destination: zone.Battlefield,
					Filter: game.Selection{
						RequiredTypes: []types.Card{types.Land},
						Supertypes:    []types.Super{types.Basic},
					},
				},
				Amount: game.Fixed(1),
			},
		}}}.Ability()),
		StaticAbilities: []game.StaticAbility{
			suspendFamilyAbility(cost.Mana{cost.G}, 2),
		},
	}}
}

func newVisceridDeepwalker() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Viscerid Deepwalker",
		ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.U}),
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Homarid, types.Warrior},
		Colors:     []color.Color{color.Blue},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 3}),
		OracleText: visceridDeepwalkerOracleText,
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:           visceridDeepwalkerAbilityText,
			ManaCost:       opt.Val(cost.Mana{cost.U}),
			ZoneOfFunction: zone.Battlefield,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.ModifyPT{
					Object:     game.SourcePermanentReference(),
					PowerDelta: game.Fixed(1),
					Duration:   game.DurationUntilEndOfTurn,
				},
			}}}.Ability(),
		}},
		StaticAbilities: []game.StaticAbility{
			suspendFamilyAbility(cost.Mana{cost.U}, 4),
		},
	}}
}

func newKeldonHalberdier() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Keldon Halberdier",
		ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.R}),
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Human, types.Warrior},
		Colors:     []color.Color{color.Red},
		Power:      opt.Val(game.PT{Value: 4}),
		Toughness:  opt.Val(game.PT{Value: 1}),
		OracleText: keldonHalberdierOracleText,
		StaticAbilities: []game.StaticAbility{
			game.FirstStrikeStaticBody,
			suspendFamilyAbility(cost.Mana{cost.R}, 4),
		},
	}}
}

func newCorpulentCorpse() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Corpulent Corpse",
		ManaCost:   opt.Val(cost.Mana{cost.O(5), cost.B}),
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Zombie},
		Colors:     []color.Color{color.Black},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
		OracleText: corpulentCorpseOracleText,
		StaticAbilities: []game.StaticAbility{
			game.FearStaticBody,
			suspendFamilyAbility(cost.Mana{cost.B}, 5),
		},
	}}
}

func newErrantEphemeron() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Errant Ephemeron",
		ManaCost:   opt.Val(cost.Mana{cost.O(6), cost.U}),
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Illusion},
		Colors:     []color.Color{color.Blue},
		Power:      opt.Val(game.PT{Value: 4}),
		Toughness:  opt.Val(game.PT{Value: 4}),
		OracleText: errantEphemeronOracleText,
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			suspendFamilyAbility(cost.Mana{cost.O(1), cost.U}, 4),
		},
	}}
}

func newMoxTantalite() *game.CardDef {
	manaAbility := game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)
	manaAbility.Text = moxTantaliteManaText
	return &game.CardDef{CardFace: game.CardFace{
		Name:          "Mox Tantalite",
		Types:         []types.Card{types.Artifact},
		OracleText:    moxTantaliteOracleText,
		ManaAbilities: []game.ManaAbility{manaAbility},
		StaticAbilities: []game.StaticAbility{
			suspendFamilyAbility(cost.Mana{cost.O(0)}, 3),
		},
	}}
}

func newInfiltratorIlKor() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Infiltrator il-Kor",
		ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.U}),
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Kor, types.Rogue},
		Colors:     []color.Color{color.Blue},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 1}),
		OracleText: infiltratorIlKorOracleText,
		StaticAbilities: []game.StaticAbility{
			game.ShadowStaticBody,
			suspendFamilyAbility(cost.Mana{cost.O(1), cost.U}, 2),
		},
	}}
}

func newIthHighArcanist() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Ith, High Arcanist",
		ManaCost:   opt.Val(cost.Mana{cost.O(5), cost.W, cost.U}),
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Human, types.Wizard},
		Colors:     []color.Color{color.White, color.Blue},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 5}),
		OracleText: ithHighArcanistOracleText,
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            ithHighArcanistAbilityText,
			AdditionalCosts: []cost.Additional{cost.T},
			ZoneOfFunction:  zone.Battlefield,
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target attacking creature",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						CombatState:   game.CombatStateAttacking,
					}),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.Untap{Object: game.TargetPermanentReference(0)}},
					{Primitive: game.PreventDamage{Object: game.TargetPermanentReference(0), All: true, CombatOnly: true}},
					{Primitive: game.PreventDamage{Object: game.TargetPermanentReference(0), All: true, CombatOnly: true, BySource: true}},
				},
			}.Ability(),
		}},
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
			suspendFamilyAbility(cost.Mana{cost.W, cost.U}, 4),
		},
	}}
}

func newRiftSower() *game.CardDef {
	manaAbility := game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)
	manaAbility.Text = riftSowerManaText
	return &game.CardDef{CardFace: game.CardFace{
		Name:          "Rift Sower",
		ManaCost:      opt.Val(cost.Mana{cost.O(2), cost.G}),
		Types:         []types.Card{types.Creature},
		Subtypes:      []types.Sub{types.Elf, types.Druid},
		Colors:        []color.Color{color.Green},
		Power:         opt.Val(game.PT{Value: 1}),
		Toughness:     opt.Val(game.PT{Value: 3}),
		OracleText:    riftSowerOracleText,
		ManaAbilities: []game.ManaAbility{manaAbility},
		StaticAbilities: []game.StaticAbility{
			suspendFamilyAbility(cost.Mana{cost.G}, 2),
		},
	}}
}
