package utils

import "slices"

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

// スライスの要素のうち、指定した条件を満たす要素を変換して新しいスライスを作成
func SlicesFilterMap[T any, U any](s []T, f func(T) *U) []U {
	var result []U
	for _, v := range s {
		if r := f(v); r != nil {
			result = append(result, *r)
		}
	}
	return result
}

// s1 と s2 の共通要素を返す
func SlicesIntersect[T comparable](s1 []T, s2 []T) []T {
	return SlicesFilter(s1, func(v T) bool {
		return slices.Contains(s2, v)
	})
}

// s1 に含まれるが s2 に含まれない要素を返す
func SlicesDifference[T comparable](s1 []T, s2 []T) []T {
	return SlicesFilter(s1, func(v T) bool {
		return !slices.Contains(s2, v)
	})
}
