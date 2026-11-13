package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestPreventDamageOneShotSourceFilterValidation proves the one-shot
// chosen-source fields are declared only on an all-damage one-shot shield.
func TestPreventDamageOneShotSourceFilterValidation(t *testing.T) {
	t.Parallel()
	valid := PreventDamage{
		Player: ControllerReference(), All: true, OneShot: true,
		SourceTypes: []types.Card{types.Artifact}, SourceColorFromEntryChoice: true,
	}
	if err := valid.validatePrimitive(nil, false); err != nil {
		t.Fatalf("validatePrimitive() = %v, want a filtered one-shot shield accepted", err)
	}
	group := PreventDamage{Player: ControllerReference(), All: true, OneShot: true, AndControlledCreatures: true}
	if err := group.validatePrimitive(nil, false); err != nil {
		t.Fatalf("validatePrimitive() = %v, want a you-and-your-creatures shield accepted", err)
	}
	for name, invalid := range map[string]PreventDamage{
		"one-shot without All":         {Player: ControllerReference(), Amount: Fixed(2), OneShot: true},
		"types without one-shot":       {Player: ControllerReference(), All: true, SourceTypes: []types.Card{types.Land}},
		"entry color without one-shot": {Player: ControllerReference(), All: true, SourceColorFromEntryChoice: true},
		"creatures without player":     {Object: SourcePermanentReference(), All: true, OneShot: true, AndControlledCreatures: true},
		"creatures with by-source":     {Player: ControllerReference(), All: true, BySource: true, AndControlledCreatures: true},
	} {
		if err := invalid.validatePrimitive(nil, false); err == nil {
			t.Errorf("%s: validatePrimitive() = nil, want rejection", name)
		}
	}
}

// TestPreventDamageRedirectValidation proves chosen-source redirection is
// declared only on an all-damage one-shot shield with a matching destination.
func TestPreventDamageRedirectValidation(t *testing.T) {
	t.Parallel()
	valid := PreventDamage{
		Player: ControllerReference(), All: true, OneShot: true,
		Redirect: DamageRedirectToObject, RedirectObject: SourcePermanentReference(),
	}
	if err := valid.validatePrimitive(nil, false); err != nil {
		t.Fatalf("validatePrimitive() = %v, want a redirect-to-this-creature shield accepted", err)
	}
	for name, invalid := range map[string]PreventDamage{
		"redirect without one-shot": {Player: ControllerReference(), All: true, Redirect: DamageRedirectToController},
		"object without destination": {Player: ControllerReference(), All: true, OneShot: true,
			Redirect: DamageRedirectToObject},
		"object on player redirect": {Player: ControllerReference(), All: true, OneShot: true,
			Redirect: DamageRedirectToController, RedirectObject: SourcePermanentReference()},
		"object without redirect": {Player: ControllerReference(), All: true, OneShot: true,
			RedirectObject: SourcePermanentReference()},
		"redirect with by-source": {Player: ControllerReference(), All: true, OneShot: true, BySource: true,
			Redirect: DamageRedirectToDamageSourceController},
		"unknown redirect": {Player: ControllerReference(), All: true, OneShot: true, Redirect: DamageRedirect(99)},
	} {
		if err := invalid.validatePrimitive(nil, false); err == nil {
			t.Errorf("%s: validatePrimitive() = nil, want rejection", name)
		}
	}
}

// TestPreventDamageRiderValidation proves half prevention and prevented-amount
// riders are declared only on one-shot chosen-source shields whose riders the
// engine applies.
func TestPreventDamageRiderValidation(t *testing.T) {
	t.Parallel()
	shield := func() PreventDamage { return PreventDamage{Player: ControllerReference(), All: true, OneShot: true} }
	riders := func(riders ...PreventedDamageRider) PreventDamage {
		p := shield()
		p.PreventedRiders = riders
		return p
	}
	half := shield()
	half.PreventHalfRoundedDown = true
	colored := riders(PreventedDamageRiderDamageSourceController)
	colored.RiderSourceColors = []color.Color{color.Red}
	reflexive := riders(PreventedDamageRiderDamageSourceController, PreventedDamageRiderControllerDraws)
	reflexive.RidersReflexive = true
	for name, valid := range map[string]PreventDamage{
		"half": half, "gain": riders(PreventedDamageRiderControllerGainsLife),
		"exile": riders(PreventedDamageRiderControllerExilesLibraryTop), "colored": colored, "reflexive": reflexive,
	} {
		if err := valid.validatePrimitive(nil, false); err != nil {
			t.Errorf("%s: validatePrimitive() = %v, want accepted", name, err)
		}
	}

	invalid := map[string]PreventDamage{
		"riders without one-shot": {Player: ControllerReference(), All: true,
			PreventedRiders: []PreventedDamageRider{PreventedDamageRiderControllerGainsLife}},
		"draw without reflexive":   riders(PreventedDamageRiderControllerDraws),
		"unknown rider":            riders(PreventedDamageRider(99)),
		"duplicate rider":          riders(PreventedDamageRiderControllerGainsLife, PreventedDamageRiderControllerGainsLife),
		"colors without riders":    {Player: ControllerReference(), All: true, OneShot: true, RiderSourceColors: []color.Color{color.Red}},
		"reflexive without riders": {Player: ControllerReference(), All: true, OneShot: true, RidersReflexive: true},
		"half without one-shot":    {Player: ControllerReference(), All: true, PreventHalfRoundedDown: true},
		"half with creatures": {Player: ControllerReference(), All: true, OneShot: true,
			AndControlledCreatures: true, PreventHalfRoundedDown: true},
		"half with redirect": {Player: ControllerReference(), All: true, OneShot: true, PreventHalfRoundedDown: true,
			Redirect: DamageRedirectToController},
		"riders with redirect": {Player: ControllerReference(), All: true, OneShot: true, Redirect: DamageRedirectToController,
			PreventedRiders: []PreventedDamageRider{PreventedDamageRiderControllerGainsLife}},
	}
	exileReflexive := riders(PreventedDamageRiderControllerExilesLibraryTop)
	exileReflexive.RidersReflexive = true
	invalid["reflexive exile"] = exileReflexive
	for name, p := range invalid {
		if err := p.validatePrimitive(nil, false); err == nil {
			t.Errorf("%s: validatePrimitive() = nil, want rejection", name)
		}
	}
}
