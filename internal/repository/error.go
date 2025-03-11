package repository

import "fmt"

type DuplicateError struct {
	Err  error
	Code string
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("db error code %s: %v", e.Code, e.Err)
}
func NewDuplicateError(code string, err error) error {
	return &DuplicateError{
		Code: code,
		Err:  err,
	}
}

type NotFoundError struct {
	Number string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("order not found: %s", e.Number)
}
func NewNotFoundError(number string) error {
	return &NotFoundError{
		Number: number,
	}
}

type ShouldBePositiveError struct {
	Err  error
	Code string
}

func (e *ShouldBePositiveError) Error() string {
	return fmt.Sprintf("db error code %s: %v", e.Code, e.Err)
}
func NewShouldBePositiveError(code string, err error) error {
	return &ShouldBePositiveError{
		Code: code,
		Err:  err,
	}
}
