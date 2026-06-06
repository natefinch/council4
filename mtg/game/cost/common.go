package cost

// Tap represents the cost of tapping a permanent.
var Tap = []Additional{{Kind: AdditionalTap}}

// T represents the cost of tapping a permanent. This is separate from Tap above, so that it can be used in contexts where you combined multiple costs.
var T = Additional{Kind: AdditionalTap}
