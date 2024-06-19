package utils

// スライスが指定した条件を満たす要素を持っているかどうか
func SlicesHas[T any](s []T, test func(T) bool) bool {
	for _, v := range s {
		if test(v) {
			return true
		}
	}
	return false
}

// スライスの要素を変換して新しいスライスを作成
func SlicesMap[T any, U any](s []T, f func(T) U) []U {
	var result []U
	for _, v := range s {
		result = append(result, f(v))
	}
	return result
}

// スライスの要素のうち、指定した条件を満たす要素の数を返す
func SlicesCount[T any](s []T, test func(T) bool) int {
	count := 0
	for _, v := range s {
		if test(v) {
			count++
		}
	}
	return count
}

// スライスの要素のうち、指定した条件を満たす要素を返す
func SlicesFilter[T any](s []T, test func(T) bool) []T {
	var result []T
	for _, v := range s {
		if test(v) {
			result = append(result, v)
		}
	}
	return result
}
