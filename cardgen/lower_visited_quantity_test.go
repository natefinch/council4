package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type visitedQuantityCase struct {
	name   string
	build  func(game.Quantity) game.Primitive
	amount func(*testing.T, game.Primitive) game.Quantity
}

func visitedQuantityCases() []visitedQuantityCase {
	object := game.TargetPermanentReference(0)
	player := game.TargetPlayerReference(0)
	return []visitedQuantityCase{
		{"Damage", func(q game.Quantity) game.Primitive {
			return game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: q}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.Damage](t, p).Amount
		}},
		{"AddCounter", func(q game.Quantity) game.Primitive { return game.AddCounter{Object: object, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.AddCounter](t, p).Amount
			}},
		{"RemoveCounter", func(q game.Quantity) game.Primitive { return game.RemoveCounter{Object: object, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.RemoveCounter](t, p).Amount
			}},
		{"AddPlayerCounter", func(q game.Quantity) game.Primitive { return game.AddPlayerCounter{Player: player, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.AddPlayerCounter](t, p).Amount
			}},
		{"MoveCounters", func(q game.Quantity) game.Primitive { return game.MoveCounters{Object: object, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.MoveCounters](t, p).Amount
			}},
		{"ModifyPT.Power", func(q game.Quantity) game.Primitive { return game.ModifyPT{Object: object, PowerDelta: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.ModifyPT](t, p).PowerDelta
			}},
		{"ModifyPT.Toughness", func(q game.Quantity) game.Primitive { return game.ModifyPT{Object: object, ToughnessDelta: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.ModifyPT](t, p).ToughnessDelta
			}},
		{"Untap", func(q game.Quantity) game.Primitive { return game.Untap{Object: object, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.Untap](t, p).Amount
			}},
		{"MovePermanent", func(q game.Quantity) game.Primitive {
			return game.MovePermanent{Object: object, Destination: zone.Exile, Amount: q}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.MovePermanent](t, p).Amount
		}},
		{"CopyStackObject", func(q game.Quantity) game.Primitive {
			return game.CopyStackObject{Object: game.TargetStackObjectReference(0), DynamicCount: q}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.CopyStackObject](t, p).DynamicCount
		}},
		{"Draw", func(q game.Quantity) game.Primitive { return game.Draw{Player: player, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.Draw](t, p).Amount
			}},
		{"Discard", func(q game.Quantity) game.Primitive { return game.Discard{Player: player, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.Discard](t, p).Amount
			}},
		{"MoveTopOfLibrary", func(q game.Quantity) game.Primitive {
			return game.MoveTopOfLibrary{Player: player, Amount: q, Destination: zone.Graveyard}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.MoveTopOfLibrary](t, p).Amount
		}},
		{"Reveal", func(q game.Quantity) game.Primitive { return game.Reveal{Player: player, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.Reveal](t, p).Amount
			}},
		{"RevealTopPartition", func(q game.Quantity) game.Primitive {
			return game.RevealTopPartition{Player: player, Amount: q}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.RevealTopPartition](t, p).Amount
		}},
		{"IterativeLibraryProcess", func(q game.Quantity) game.Primitive {
			return game.IterativeLibraryProcess{Player: player, PreExile: q}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.IterativeLibraryProcess](t, p).PreExile
		}},
		{"GainLife", func(q game.Quantity) game.Primitive { return game.GainLife{Player: player, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.GainLife](t, p).Amount
			}},
		{"LoseLife", func(q game.Quantity) game.Primitive { return game.LoseLife{Player: player, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.LoseLife](t, p).Amount
			}},
		{"SacrificePermanents", func(q game.Quantity) game.Primitive {
			return game.SacrificePermanents{Player: player, Amount: q}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.SacrificePermanents](t, p).Amount
		}},
		{"ApplyContinuous.ChooseUpTo", func(q game.Quantity) game.Primitive {
			return game.ApplyContinuous{Object: opt.Val(object), ChooseUpTo: q}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.ApplyContinuous](t, p).ChooseUpTo
		}},
		{"CreateToken.Amount", func(q game.Quantity) game.Primitive { return game.CreateToken{Recipient: opt.Val(player), Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.CreateToken](t, p).Amount
			}},
		{"CreateToken.Power", func(q game.Quantity) game.Primitive {
			return game.CreateToken{Recipient: opt.Val(player), Power: opt.Val(q)}
		},
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.CreateToken](t, p).Power.Val
			}},
		{"CreateToken.Toughness", func(q game.Quantity) game.Primitive {
			return game.CreateToken{Recipient: opt.Val(player), Toughness: opt.Val(q)}
		},
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.CreateToken](t, p).Toughness.Val
			}},
		{"MoveCard", func(q game.Quantity) game.Primitive { return game.MoveCard{Player: player, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.MoveCard](t, p).Amount
			}},
		{"PreventDamage", func(q game.Quantity) game.Primitive { return game.PreventDamage{Object: object, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.PreventDamage](t, p).Amount
			}},
		{"RepeatProcess.Times", func(q game.Quantity) game.Primitive {
			return game.RepeatProcess{Times: q, Body: game.Mode{Sequence: []game.Instruction{{Primitive: game.Explore{Creature: object}}}}.Ability()}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.RepeatProcess](t, p).Times
		}},
		{"Pay.DynamicGenericManaCost", func(q game.Quantity) game.Primitive {
			dynamic := q.DynamicAmount().Val
			return game.Pay{Payment: game.ResolutionPayment{Payer: opt.Val(player), DynamicGenericManaCost: opt.Val(&dynamic)}}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return game.Dynamic(*captureTestPrimitive[game.Pay](t, p).Payment.DynamicGenericManaCost.Val)
		}},
		{"Pay.ManaCostMultiplier", func(q game.Quantity) game.Primitive {
			dynamic := q.DynamicAmount().Val
			return game.Pay{Payment: game.ResolutionPayment{Payer: opt.Val(player), ManaCostMultiplier: opt.Val(&dynamic)}}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return game.Dynamic(*captureTestPrimitive[game.Pay](t, p).Payment.ManaCostMultiplier.Val)
		}},
	}
}

func TestVisitedPrimitiveQuantitiesProjectEverySupportedField(t *testing.T) {
	t.Parallel()
	player := game.TargetPlayerReference(0)
	for _, tc := range visitedQuantityCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.name != "Pay.DynamicGenericManaCost" && tc.name != "Pay.ManaCostMultiplier" {
				projected, ok := rebaseTargetedPrimitive(tc.build(game.Fixed(2)), 2, 1)
				if !ok || tc.amount(t, projected).IsDynamic() || tc.amount(t, projected).Value() != 2 {
					t.Fatal("projection changed a fixed quantity")
				}
			}
			for _, amount := range []game.DynamicAmount{
				{Kind: game.DynamicAmountObjectPower, Object: game.TargetPermanentReference(0)},
				{Kind: game.DynamicAmountCountCardsInZone, Player: &player, CardZone: zone.Hand, Multiplier: 1, Selection: new(game.Selection)},
				{Kind: game.DynamicAmountCountSelector, Group: game.PlayerControlledGroup(player, game.Selection{}), Multiplier: 1},
				{Kind: game.DynamicAmountMaxOf, Operands: []game.DynamicAmount{{
					Kind: game.DynamicAmountCountCardsInZone, Player: &player, CardZone: zone.Hand, Multiplier: 1, Selection: new(game.Selection),
				}}},
			} {
				original := tc.build(game.Dynamic(amount))
				projected, ok := rebaseTargetedPrimitive(original, 2, 1)
				if !ok {
					t.Fatal("supported quantity field was not projectable")
				}
				got := tc.amount(t, projected).DynamicAmount().Val
				if got.Object.Kind() == game.ObjectReferenceTargetPermanent && got.Object.TargetIndex() != 2 ||
					got.Player != nil && *got.Player != game.TargetPlayerReference(2) ||
					len(got.Operands) != 0 && *got.Operands[0].Player != game.TargetPlayerReference(2) {
					t.Fatal("body primitive amount retained a clause-local target index")
				}
				if anchor, exists := got.Group.PlayerAnchor(); exists && anchor != game.TargetPlayerReference(2) {
					t.Fatal("body primitive amount retained a clause-local group anchor")
				}
				before := tc.amount(t, original).DynamicAmount().Val
				if before.Player != nil && *before.Player != player ||
					len(before.Operands) != 0 && *before.Operands[0].Player != player {
					t.Fatal("projection mutated the original quantity")
				}
				second, ok := rebaseTargetedPrimitive(original, 3, 1)
				if !ok || tc.amount(t, second).DynamicAmount().Val.Player != nil &&
					*tc.amount(t, second).DynamicAmount().Val.Player != game.TargetPlayerReference(3) {
					t.Fatal("reused definition did not build an independent second projection")
				}
				if got.Player != nil && *got.Player != game.TargetPlayerReference(2) ||
					len(got.Operands) != 0 && *got.Operands[0].Player != game.TargetPlayerReference(2) {
					t.Fatal("second projection mutated the first")
				}
			}
			unscoped := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountCountSelector,
				Group: game.PlayerGroupControlledGroup(game.TargetedPlayersReference(), game.Selection{}), Multiplier: 1})
			if _, ok := rebaseTargetedPrimitive(tc.build(unscoped), 2, 1); ok {
				t.Fatal("whole-target-group quantity bypassed target-spec scope")
			}
		})
	}
}
