package main

// this is a little misleading, but done this way to simplify producer logic
// dataPopulator uses start and end fields in both ways, as indices and actual values
// in another implementation this could be split into start and end indices and readData can return another type of value
// but for the sake of simplicity it just returns current index as retrieved value
type dataPopulator struct {
	start   int
	end     int
	current int
}

func newDataPopulator(start, end int) dataPopulator {
	return dataPopulator{
		start:   start,
		end:     end,
		current: start,
	}
}

func (dp *dataPopulator) readData() (int, bool) {
	if dp.current == dp.end {
		return 0, false
	}

	cur := dp.current
	dp.current++
	return cur, true
}
