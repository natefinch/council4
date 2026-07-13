package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// targetBearingPrimitive describes one target-bearing primitive variant so the
// completeness test below can assert the shared transformPrimitiveTargetIndices
// walker both rebases (uniform offset) and remaps (lookup table) it in the
// correct numbering domain. Adding a new target-bearing primitive kind to the
// walker without adding a row here leaves the kind unguarded, so the table is
// the single inventory that locks the walker's coverage.
type targetBearingPrimitive struct {
	name string
	// build returns a primitive carrying a single clause-local target index 0 in
	// the named numbering domain.
	build func() game.Primitive
	// domain is the numbering domain of the carried index: object indices share
	// the global target list, card indices are counted among card targets only.
	domain targetIndexKind
	// index extracts the (single) target index the walker rewrote, reporting
	// false if the primitive carries none.
	index func(game.Primitive) (int, bool)
}

// objectIndex pulls the single object target index out of an object-targeting
// primitive via assertion chains (the codebase avoids type switches).
func objectIndex(p game.Primitive) (int, bool) {
	if v, ok := p.(game.Destroy); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.AddCounter); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.MoveCounters); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.ModifyPT); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.Fight); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.Tap); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.TapOrUntap); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.SkipNextUntap); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.Untap); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.RemoveFromCombat); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.Exile); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.Bounce); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.CounterObject); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.CopyStackObject); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.ChooseNewTargets); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.Regenerate); ok {
		return v.Object.TargetIndex(), true
	}
	if v, ok := p.(game.Attach); ok {
		return v.Target.TargetIndex(), true
	}
	if v, ok := p.(game.PreventDamage); ok {
		return v.Object.TargetIndex(), true
	}
	return 0, false
}

// playerIndex pulls the single player target index out of a player-targeting
// primitive via assertion chains.
func playerIndex(p game.Primitive) (int, bool) {
	if v, ok := p.(game.AddPlayerCounter); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.Draw); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.Discard); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.Mill); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.ExileTopOfLibrary); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.RevealUntil); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.GainLife); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.LoseLife); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.SacrificePermanents); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.LookAtHand); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.BecomeMonarch); ok {
		return v.Player.TargetIndex(), true
	}
	if v, ok := p.(game.MoveCard); ok {
		return v.Player.TargetIndex(), true
	}
	return 0, false
}

func objectPrimitive(name string, build func() game.Primitive) targetBearingPrimitive {
	return targetBearingPrimitive{name: name, build: build, domain: targetIndexObject, index: objectIndex}
}

func playerPrimitive(name string, build func() game.Primitive) targetBearingPrimitive {
	return targetBearingPrimitive{name: name, build: build, domain: targetIndexObject, index: playerIndex}
}

func targetBearingPrimitives() []targetBearingPrimitive {
	obj := func() game.ObjectReference { return game.TargetPermanentReference(0) }
	plr := func() game.PlayerReference { return game.TargetPlayerReference(0) }
	cards := []targetBearingPrimitive{
		objectPrimitive("Destroy", func() game.Primitive { return game.Destroy{Object: obj()} }),
		objectPrimitive("AddCounter", func() game.Primitive { return game.AddCounter{Object: obj()} }),
		objectPrimitive("MoveCounters", func() game.Primitive {
			return game.MoveCounters{Object: obj(), Source: game.CounterSourceSpec{Kind: game.CounterSourceSelf}}
		}),
		objectPrimitive("ModifyPT", func() game.Primitive { return game.ModifyPT{Object: obj()} }),
		objectPrimitive("Fight", func() game.Primitive { return game.Fight{Object: obj(), RelatedObject: obj()} }),
		objectPrimitive("Tap", func() game.Primitive { return game.Tap{Object: obj()} }),
		objectPrimitive("TapOrUntap", func() game.Primitive { return game.TapOrUntap{Object: obj()} }),
		objectPrimitive("SkipNextUntap", func() game.Primitive { return game.SkipNextUntap{Object: obj()} }),
		objectPrimitive("Untap", func() game.Primitive { return game.Untap{Object: obj()} }),
		objectPrimitive("RemoveFromCombat", func() game.Primitive { return game.RemoveFromCombat{Object: obj()} }),
		objectPrimitive("Exile", func() game.Primitive { return game.Exile{Object: obj()} }),
		objectPrimitive("Bounce", func() game.Primitive { return game.Bounce{Object: obj()} }),
		objectPrimitive("CounterObject", func() game.Primitive { return game.CounterObject{Object: obj()} }),
		objectPrimitive("CopyStackObject", func() game.Primitive {
			return game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true}
		}),
		objectPrimitive("ChooseNewTargets", func() game.Primitive {
			return game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)}
		}),
		objectPrimitive("Regenerate", func() game.Primitive { return game.Regenerate{Object: obj()} }),
		objectPrimitive("Attach", func() game.Primitive {
			return game.Attach{Attachment: game.SourcePermanentReference(), Target: obj()}
		}),
		objectPrimitive("PreventDamage", func() game.Primitive { return game.PreventDamage{Object: obj()} }),
		playerPrimitive("LookAtHand", func() game.Primitive { return game.LookAtHand{Player: plr()} }),
		playerPrimitive("BecomeMonarch", func() game.Primitive { return game.BecomeMonarch{Player: plr()} }),
		playerPrimitive("AddPlayerCounter", func() game.Primitive { return game.AddPlayerCounter{Player: plr()} }),
		playerPrimitive("Draw", func() game.Primitive { return game.Draw{Player: plr()} }),
		playerPrimitive("Discard", func() game.Primitive { return game.Discard{Player: plr()} }),
		playerPrimitive("Mill", func() game.Primitive { return game.Mill{Player: plr()} }),
		playerPrimitive("ExileTopOfLibrary", func() game.Primitive { return game.ExileTopOfLibrary{Player: plr()} }),
		playerPrimitive("RevealUntil", func() game.Primitive { return game.RevealUntil{Player: plr()} }),
		playerPrimitive("GainLife", func() game.Primitive { return game.GainLife{Player: plr()} }),
		playerPrimitive("LoseLife", func() game.Primitive { return game.LoseLife{Player: plr()} }),
		playerPrimitive("SacrificePermanents", func() game.Primitive { return game.SacrificePermanents{Player: plr()} }),
		playerPrimitive("MoveCard player-zone", func() game.Primitive {
			return game.MoveCard{Player: plr(), FromZone: zone.Graveyard, Destination: zone.Exile}
		}),
		{
			name: "Damage any-target",
			build: func() game.Primitive {
				return game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}
			},
			domain: targetIndexObject,
			index: func(p game.Primitive) (int, bool) {
				damage, ok := p.(game.Damage)
				if !ok {
					return 0, false
				}
				ref, ok := damage.Recipient.AnyTargetObjectReference()
				if !ok {
					return 0, false
				}
				return ref.TargetIndex(), true
			},
		},
		{
			name:   "ApplyContinuous",
			build:  func() game.Primitive { return game.ApplyContinuous{Object: opt.Val(obj())} },
			domain: targetIndexObject,
			index: func(p game.Primitive) (int, bool) {
				apply, ok := p.(game.ApplyContinuous)
				if !ok || !apply.Object.Exists {
					return 0, false
				}
				return apply.Object.Val.TargetIndex(), true
			},
		},
		{
			name:   "ApplyRule",
			build:  func() game.Primitive { return game.ApplyRule{Object: opt.Val(obj())} },
			domain: targetIndexObject,
			index: func(p game.Primitive) (int, bool) {
				apply, ok := p.(game.ApplyRule)
				if !ok || !apply.Object.Exists {
					return 0, false
				}
				return apply.Object.Val.TargetIndex(), true
			},
		},
		{
			name:   "CreateToken",
			build:  func() game.Primitive { return game.CreateToken{Recipient: opt.Val(plr())} },
			domain: targetIndexObject,
			index: func(p game.Primitive) (int, bool) {
				token, ok := p.(game.CreateToken)
				if !ok || !token.Recipient.Exists {
					return 0, false
				}
				return token.Recipient.Val.TargetIndex(), true
			},
		},
		{
			name: "MoveCard single-card",
			build: func() game.Primitive {
				return game.MoveCard{Card: game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0}, FromZone: zone.Graveyard, Destination: zone.Exile}
			},
			domain: targetIndexCard,
			index: func(p game.Primitive) (int, bool) {
				move, ok := p.(game.MoveCard)
				if !ok || move.Card.Kind != game.CardReferenceTarget {
					return 0, false
				}
				return move.Card.TargetIndex, true
			},
		},
		{
			name: "PutOnBattlefield card source",
			build: func() game.Primitive {
				return game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0})}
			},
			domain: targetIndexCard,
			index: func(p game.Primitive) (int, bool) {
				put, ok := p.(game.PutOnBattlefield)
				if !ok {
					return 0, false
				}
				card, ok := put.Source.CardRef()
				if !ok || card.Kind != game.CardReferenceTarget {
					return 0, false
				}
				return card.TargetIndex, true
			},
		},
	}
	return cards
}

// TestTransformPrimitiveTargetIndicesCoversEveryKind proves the single shared
// walker rebases (uniform offset) and remaps (lookup table) every target-bearing
// primitive kind in the correct numbering domain. Object-domain indices use the
// global target offset/table; card-domain indices use the separate card offset
// and fail closed under the remap table, which cannot express card numbering.
func TestTransformPrimitiveTargetIndicesCoversEveryKind(t *testing.T) {
	t.Parallel()
	const objectOffset, cardOffset = 5, 7
	for _, tc := range targetBearingPrimitives() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rebased, ok := rebaseTargetedPrimitive(tc.build(), objectOffset, cardOffset)
			if !ok {
				t.Fatalf("%s was not rebased", tc.name)
			}
			gotIdx, found := tc.index(rebased)
			if !found {
				t.Fatalf("%s carried no rewritten target index after rebase", tc.name)
			}
			wantRebased := objectOffset
			if tc.domain == targetIndexCard {
				wantRebased = cardOffset
			}
			if gotIdx != wantRebased {
				t.Fatalf("%s rebased index = %d, want %d", tc.name, gotIdx, wantRebased)
			}

			seq := []game.Instruction{{Primitive: tc.build()}}
			remapOK := remapTargetedSequence(seq, []int{9})
			if tc.domain == targetIndexCard {
				if remapOK {
					t.Fatalf("%s card-domain index was remapped; the lookup table cannot express card numbering and must fail closed", tc.name)
				}
				return
			}
			if !remapOK {
				t.Fatalf("%s object-domain index was not remapped", tc.name)
			}
			gotIdx, found = tc.index(seq[0].Primitive)
			if !found {
				t.Fatalf("%s carried no rewritten target index after remap", tc.name)
			}
			if gotIdx != 9 {
				t.Fatalf("%s remapped index = %d, want 9", tc.name, gotIdx)
			}
		})
	}
}

// TestTransformPassThroughPrimitives proves the walker accepts the allowlisted
// primitives that carry no directly-rewritten target index, leaving them
// unchanged under both rebase and remap rather than failing closed.
func TestTransformPassThroughPrimitives(t *testing.T) {
	t.Parallel()
	delayed := game.CreateDelayedTrigger{}
	rebased, ok := rebaseTargetedPrimitive(delayed, 5, 7)
	if !ok {
		t.Fatal("CreateDelayedTrigger was not passed through by rebase")
	}
	if _, ok := rebased.(game.CreateDelayedTrigger); !ok {
		t.Fatalf("rebased pass-through primitive = %+v", rebased)
	}
	seq := []game.Instruction{{Primitive: delayed}}
	if !remapTargetedSequence(seq, []int{9}) {
		t.Fatal("CreateDelayedTrigger was not passed through by remap")
	}
}

// TestTransformMixedObjectAndCardSequence proves a single clause carrying both a
// non-card object target and a later card target rebases each in its own
// numbering domain: the object index shifts by the global offset while the card
// index shifts by the (smaller) card offset, exactly as the two coincident-base
// callers require.
func TestTransformMixedObjectAndCardSequence(t *testing.T) {
	t.Parallel()
	seq := []game.Instruction{
		{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.MoveCard{Card: game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0}, FromZone: zone.Graveyard, Destination: zone.Exile}},
	}
	if !rebaseTargetedSequence(seq, 2, 1) {
		t.Fatal("mixed object+card sequence was not rebased")
	}
	destroy, ok := seq[0].Primitive.(game.Destroy)
	if !ok || destroy.Object.TargetIndex() != 2 {
		t.Fatalf("rebased object target = %+v, want index 2", seq[0].Primitive)
	}
	move, ok := seq[1].Primitive.(game.MoveCard)
	if !ok || move.Card.TargetIndex != 1 {
		t.Fatalf("rebased card target = %+v, want index 1", seq[1].Primitive)
	}
}
