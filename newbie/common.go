package newbie

// スライスが指定した条件を満たす要素を持っているかどうか
func slicesHas[T any](s []T, test func(T) bool) bool {
	for _, v := range s {
		if test(v) {
			return true
		}
	}
	return false
}
