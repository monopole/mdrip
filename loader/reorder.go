package loader

import "github.com/monopole/mdrip/model"

func shiftToTop(x []model.Tutorial, top string) []model.Tutorial {
	var result []model.Tutorial
	var other []model.Tutorial
	for _, f := range x {
		if f.Name() == top {
			result = append(result, f)
		} else {
			other = append(other, f)
		}
	}
	return append(result, other...)
}

// reorder tutorial array in some fashion
func reorder(x []model.Tutorial, ordering []string) []model.Tutorial {
	for i := len(ordering) - 1; i >= 0; i-- {
		x = shiftToTop(x, ordering[i])
	}
	return shiftToTop(x, "README")
}
