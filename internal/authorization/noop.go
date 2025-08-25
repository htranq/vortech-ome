package authorization

import "errors"

type noop struct {
}

func (*noop) Sign(_ string) string {
	return ""
}

func (*noop) Verify(_, _ string) error {
	return errors.New("cannot verify the signature with noop")
}
