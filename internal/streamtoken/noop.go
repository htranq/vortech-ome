package streamtoken

import "errors"

type noop struct {
}

func (*noop) Issue(_ string) (string, error) {
	return "", errors.New("unsupported")
}

func (*noop) Verify(_ string) (bool, error) {
	return false, errors.New("unsupported")
}
