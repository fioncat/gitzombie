package common

type Args []string

func (a Args) GetDefault(idx int, def string) string {
	if idx < 0 || idx >= len(a) {
		return def
	}
	return a[idx]
}

func (a Args) Get(idx int) string { return a.GetDefault(idx, "") }
