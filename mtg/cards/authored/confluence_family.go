package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

var FieryConfluence = cardset.Entry{
	OracleID: "3c22e031-4804-4c31-bd3c-c3f29d456b34",
	Name:     "Fiery Confluence",
	New:      newFieryConfluence,
	Verification: "mtg/rules/authored_confluences_test.go: " +
		"TestFieryConfluenceRepeatsOpponentDamageAndDestroysArtifact",
}

func newFieryConfluence() *game.CardDef {
	const text = "Choose three. You may choose the same mode more than once.\n" +
		"• Fiery Confluence deals 1 damage to each creature.\n" +
		"• Fiery Confluence deals 2 damage to each opponent.\n" +
		"• Destroy target artifact."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Fiery Confluence",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.R, cost.R}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Red},
			OracleText: text,
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{
						Text: "Fiery Confluence deals 1 damage to each creature.",
						Sequence: []game.Instruction{{Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						}}},
					},
					{
						Text: "Fiery Confluence deals 2 damage to each opponent.",
						Sequence: []game.Instruction{{Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
						}}},
					},
					{
						Text: "Destroy target artifact.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target artifact",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
						}},
						Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
					},
				},
				MinModes:            3,
				MaxModes:            3,
				AllowDuplicateModes: true,
			}),
		},
	}
}

var MysticConfluence = cardset.Entry{
	OracleID: "cd11c27b-9368-4622-bc5a-2e0b993cc42b",
	Name:     "Mystic Confluence",
	New:      newMysticConfluence,
	Verification: "mtg/rules/authored_confluences_test.go: " +
		"TestMysticConfluencePaysForCounterAndDrawsTwice",
}

func newMysticConfluence() *game.CardDef {
	const text = "Choose three. You may choose the same mode more than once.\n" +
		"• Counter target spell unless its controller pays {3}.\n" +
		"• Return target creature to its owner's hand.\n" +
		"• Draw a card."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Mystic Confluence",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.U, cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			OracleText: text,
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{
						Text: "Counter target spell unless its controller pays {3}.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target spell",
							Allow:      game.TargetAllowStackObject,
							Predicate:  game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
						}},
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{Payment: game.ResolutionPayment{
									Prompt:   "Pay {3}?",
									Payer:    opt.Val(game.AffectedTargetControllerReference(0)),
									ManaCost: opt.Val(cost.Mana{cost.O(3)}),
								}},
								PublishResult: "paid",
							},
							{
								Primitive:  game.CounterObject{Object: game.TargetStackObjectReference(0)},
								ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriFalse}),
							},
						},
					},
					{
						Text: "Return target creature to its owner's hand.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						}},
						Sequence: []game.Instruction{{Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						}}},
					},
					{
						Text: "Draw a card.",
						Sequence: []game.Instruction{{Primitive: game.Draw{
							Player: game.ControllerReference(),
							Amount: game.Fixed(1),
						}}},
					},
				},
				MinModes:            3,
				MaxModes:            3,
				AllowDuplicateModes: true,
			}),
		},
	}
}

var RighteousConfluence = cardset.Entry{
	OracleID: "ee28937b-7ead-45ad-8448-35c51a09fb32",
	Name:     "Righteous Confluence",
	New:      newRighteousConfluence,
	Programs: []cardprogram.Program{{
		Reference: righteousConfluenceKnightProgram,
		Resolve:   resolveRighteousConfluenceKnight,
	}},
	Verification: "mtg/rules/authored_confluences_test.go: " +
		"TestRighteousConfluenceCreatesTwoVigilantKnightsAndExilesEnchantment",
}

var righteousConfluenceKnightProgram = game.ProgramReference{
	ID:      "ee28937b-7ead-45ad-8448-35c51a09fb32/spell/0",
	Version: "1",
}

func newRighteousConfluence() *game.CardDef {
	const text = "Choose three. You may choose the same mode more than once.\n" +
		"• Create a 2/2 white Knight creature token with vigilance.\n" +
		"• Exile target enchantment.\n" +
		"• You gain 5 life."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Righteous Confluence",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.W, cost.W}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.White},
			OracleText: text,
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{
						Text:    "Create a 2/2 white Knight creature token with vigilance.",
						Program: righteousConfluenceKnightProgram,
					},
					{
						Text: "Exile target enchantment.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target enchantment",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
						}},
						Sequence: []game.Instruction{{Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						}}},
					},
					{
						Text: "You gain 5 life.",
						Sequence: []game.Instruction{{Primitive: game.GainLife{
							Amount: game.Fixed(5),
							Player: game.ControllerReference(),
						}}},
					},
				},
				MinModes:            3,
				MaxModes:            3,
				AllowDuplicateModes: true,
			}),
		},
	}
}

func resolveRighteousConfluenceKnight(ctx cardprogram.Context) error {
	_, err := ctx.CreateTokens(righteousConfluenceKnightToken(), cardprogram.TokenRecipient{
		Player: ctx.View().Source().Controller,
		Count:  1,
	})
	return err
}

func righteousConfluenceKnightToken() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Knight Token",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Knight},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
	}}
}

var WretchedConfluence = cardset.Entry{
	OracleID: "fb430640-81b6-4720-84d6-b8ec3af6a2c3",
	Name:     "Wretched Confluence",
	New:      newWretchedConfluence,
	Verification: "mtg/rules/authored_confluences_test.go: " +
		"TestWretchedConfluenceRepeatsPlayerModeAndReturnsCreatureCard",
}

func newWretchedConfluence() *game.CardDef {
	const text = "Choose three. You may choose the same mode more than once.\n" +
		"• Target player draws a card and loses 1 life.\n" +
		"• Target creature gets -2/-2 until end of turn.\n" +
		"• Return target creature card from your graveyard to your hand."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Wretched Confluence",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.B, cost.B}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Black},
			OracleText: text,
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{
						Text: "Target player draws a card and loses 1 life.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target player",
							Allow:      game.TargetAllowPlayer,
						}},
						Sequence: []game.Instruction{
							{Primitive: game.Draw{Player: game.TargetPlayerReference(0), Amount: game.Fixed(1)}},
							{Primitive: game.LoseLife{Player: game.TargetPlayerReference(0), Amount: game.Fixed(1)}},
						},
					},
					{
						Text: "Target creature gets -2/-2 until end of turn.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						}},
						Sequence: []game.Instruction{{Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(-2),
							ToughnessDelta: game.Fixed(-2),
							Duration:       game.DurationUntilEndOfTurn,
						}}},
					},
					{
						Text: "Return target creature card from your graveyard to your hand.",
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature card from your graveyard",
							Allow:      game.TargetAllowCard,
							TargetZone: zone.Graveyard,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								Controller:    game.ControllerYou,
							}),
						}},
						Sequence: []game.Instruction{{Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						}}},
					},
				},
				MinModes:            3,
				MaxModes:            3,
				AllowDuplicateModes: true,
			}),
		},
	}
}
