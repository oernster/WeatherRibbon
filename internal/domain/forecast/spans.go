package forecast

import "time"

// span is one block fixed to the instant it starts.
type span struct {
	start time.Time
	block Period
}

// rainSpans answers the blocks a day's rain is summed over: blocks that never overlap, so no hour's
// rain is counted twice (FR-408). At each step it takes the block reaching exactly to the next step
// where the step carries one (the 1-hour block while steps are hourly, the 6-hour block once they are
// six hours apart), else its shortest block giving rain. A block starting before the last one taken
// ends is passed over.
func rainSpans(steps []Step) []span {
	var taken []span
	var covered time.Time
	for index, each := range steps {
		gap := time.Duration(0)
		if next := index + 1; next < len(steps) {
			gap = steps[next].Time.Sub(each.Time)
		}
		block, found := rainBlock(each, gap)
		if !found || each.Time.Before(covered) {
			continue
		}
		taken = append(taken, span{start: each.Time, block: block})
		covered = each.Time.Add(block.length())
	}
	return taken
}

// rainBlock answers the block of step to take rain from, given the gap to the next step: the block
// giving rain that spans the gap exactly, else the shortest giving rain; false when none gives rain.
func rainBlock(step Step, gap time.Duration) (Period, bool) {
	var giving []Period
	for _, each := range step.periods() {
		if each.HasRain {
			giving = append(giving, each)
		}
	}
	if len(giving) == 0 {
		return Period{}, false
	}
	for _, each := range giving {
		if each.length() == gap {
			return each, true
		}
	}
	return giving[0], true
}

// blockSpans answers every block every step carries, fixed to its step's instant.
func blockSpans(steps []Step) []span {
	var all []span
	for _, each := range steps {
		for _, block := range each.periods() {
			all = append(all, span{start: each.Time, block: block})
		}
	}
	return all
}
