package common

func MustVal[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func NotNil(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
