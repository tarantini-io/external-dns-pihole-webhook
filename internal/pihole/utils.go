package pihole

func Map[S any, D any](source []S, dest *[]D, mapFun func(S) (D, error)) {
	if cap(*dest) == 0 {
		*dest = []D{}
	}
	for _, s := range source {
		mapped, err := mapFun(s)
		if err != nil {
			continue
		}
		*dest = append(*dest, mapped)
	}
}
