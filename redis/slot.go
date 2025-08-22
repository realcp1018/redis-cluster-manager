package redis

import (
	"fmt"
	"strconv"
	"strings"
)

type SlotRange struct {
	Start     int // start of the slot range
	End       int // end of the slot range
	SlotCount int
}

// newSlotRanges creates a slice of SlotRange from a string like "1 2-100 101-200"
func newSlotRanges(slotStr string) []*SlotRange {
	var slotRanges []*SlotRange
	if len(slotStr) == 0 {
		return nil
	}
	slots := strings.Split(slotStr, " ")
	for _, slot := range slots {
		if strings.Contains(slot, "-") {
			// it's a range
			parts := strings.Split(slot, "-")
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			slotRanges = append(slotRanges, &SlotRange{
				Start:     start,
				End:       end,
				SlotCount: end - start + 1,
			})
		} else {
			// it's a single slot
			slotNum, _ := strconv.Atoi(slot)
			slotRanges = append(slotRanges, &SlotRange{
				Start:     slotNum,
				End:       slotNum,
				SlotCount: 1,
			})
		}
	}
	return slotRanges
}

func (s *SlotRange) ContainsSlot(slot int) bool {
	return slot >= s.Start && slot <= s.End
}

func (s *SlotRange) String() string {
	return fmt.Sprintf("[%d-%d]", s.Start, s.End)
}
