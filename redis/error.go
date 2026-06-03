package redis

import "strings"

const loadingErrorMessage = "LOADING Redis is loading the dataset in memory"

func IsLoadingError(err error) bool {
	return err != nil && strings.Contains(err.Error(), loadingErrorMessage)
}
