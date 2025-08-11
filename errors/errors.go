package errors

import (
	"errors"
	"fmt"
)

func New(text string) error {
	return errors.New(text)
}

func Newf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", text, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", msg, err)
}
