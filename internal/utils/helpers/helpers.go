package helpers

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
)

func GetCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	strs := strings.Split((runtime.FuncForPC(pc).Name()), "/")
	return strs[len(strs)-1]
}

func EntityName(entity any) string {
	return reflect.TypeOf(entity).Name()
}

func Filter[T, T2 any](collection []T, object T2, inclusionTest func(T, T2, int) bool) []T {
	result := make([]T, 0)

	for i, item := range collection {
		if inclusionTest(item, object, i) {
			result = append(result, item)
		}
	}

	return result
}

func Find[T, T2 any](collection []T, object T2, inclusionTest func(T, T2, int) bool) *int {
	for i, item := range collection {
		if inclusionTest(item, object, i) {
			return &i
		}
	}

	return nil
}

func FindIndexes[T, T2 any](collection []T, object T2, inclusionTest func(T, T2, int) bool) []int {
	result := make([]int, 0)

	for i, item := range collection {
		if inclusionTest(item, object, i) {
			result = append(result, i)
		}
	}

	return result
}

func Remove[T any](slice []T, s int) []T {
	if len(slice) <= 1 {
		return make([]T, 0)
	} else {
		return append(slice[:s], slice[s+1:]...)
	}
}

func CloneDeep[T any](a T) (res T, err error) {
	b, err := json.Marshal(a)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(b, &res)
	return res, err
}
