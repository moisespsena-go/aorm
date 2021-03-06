package aorm

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	// ErrRecordNotFound record not found error, happens when haven'T find any matched data when looking up with a struct
	ErrRecordNotFound = errors.New("record not found")
	// ErrInvalidSQL invalid SQL error, happens when you passed invalid SQL
	ErrInvalidSQL = errors.New("invalid SQL")
	// ErrInvalidTransaction invalid transaction when you are trying to `Commit` or `Rollback`
	ErrInvalidTransaction = errors.New("no valid transaction")
	// ErrCantStartTransaction can'T start transaction when you are trying to start one with `Begin`
	ErrCantStartTransaction = errors.New("can'T start transaction")
	// ErrUnaddressable unaddressable value
	ErrUnaddressable = errors.New("using unaddressable value")
	// ErrSingleUpdateKey single UPDATE require primary key value
	ErrSingleUpdateKey = errors.New("Single UPDATE require primary key value.")
)

// Errors contains all happened errors
type Errors []error

func WalkErr(cb func(err error) (stop bool), errs ...error) (stop bool) {
	for _, err := range errs {
		if err == nil {
			continue
		}

		if cb(err) {
			return true
		}

		if err, ok := err.(interface{ Err() error }); ok {
			if WalkErr(cb, err.Err()) {
				return true
			}
		}
		if err, ok := err.(interface{ Cause() error }); ok {
			if WalkErr(cb, err.Cause()) {
				return true
			}
		}

		if errs, ok := err.(Errors); ok {
			if WalkErr(cb, errs...) {
				return true
			}
		} else if errs, ok := err.(interface{ Errors() []error }); ok {
			if WalkErr(cb, errs.Errors()...) {
				return true
			}
		} else if errs, ok := err.(interface{ GetErrors() []error }); ok {
			if WalkErr(cb, errs.GetErrors()...) {
				return true
			}
		}
	}
	return false
}

func IsError(expected error, err ...error) (is bool) {
	return WalkErr(func(err error) (stop bool) {
		return err == expected
	}, err...)
}

func ErrorByType(expected reflect.Type, err ...error) (theError error) {
	expected = indirectRealType(expected)
	WalkErr(func(err error) (stop bool) {
		if indirectRealType(reflect.TypeOf(err)) == expected {
			theError = err
			return true
		}
		return false
	}, err...)
	return
}

func IsErrorTyp(expected reflect.Type, err ...error) (is bool) {
	return ErrorByType(expected, err...) != nil
}

// IsRecordNotFoundError returns current error has record not found error or not
func IsRecordNotFoundError(err error) bool {
	return IsError(ErrRecordNotFound, err)
}

// GetErrors gets all happened errors
func (errs Errors) GetErrors() []error {
	return errs
}

// Add adds an error
func (errs Errors) Add(newErrors ...error) Errors {
	for _, err := range newErrors {
		if err == nil {
			continue
		}

		if errors, ok := err.(Errors); ok {
			errs = errs.Add(errors...)
		} else {
			ok = true
			for _, e := range errs {
				if err == e {
					ok = false
				}
			}
			if ok {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// error format happened errors
func (errs Errors) Error() string {
	var errors = []string{}
	for _, e := range errs {
		errors = append(errors, e.Error())
	}
	return strings.Join(errors, "; ")
}

// Represents Query error
type QueryError struct {
	QueryInfo
	cause error
}

func NewQueryError(cause error, q Query, varBinder func(i int) string) *QueryError {
	qi := NewQueryInfo(q, varBinder)
	return &QueryError{*qi, cause}
}

// Returns the original error
func (e *QueryError) Cause() error {
	return e.cause
}

func (e *QueryError) Error() string {
	var args = make([]interface{}, len(e.Query.Args), len(e.Query.Args))
	e.EachArgs(func(i int, name string, value interface{}) {
		if vlr, ok := value.(driver.Valuer); ok {
			if v, err := vlr.Value(); err == nil {
				value = v
			}
		}
		args[i] = sql.Named(e.argsName[i], value)
	})
	return e.cause.Error() + "\n" + (Query{e.Query.Query, args}.String())
}

func IsQueryError(err ...error) bool {
	return ErrorByType(reflect.TypeOf(QueryError{}), err...) != nil
}

func GetQueryError(err ...error) *QueryError {
	if result := ErrorByType(reflect.TypeOf(QueryError{}), err...); result != nil {
		return result.(*QueryError)
	}
	return nil
}

type DuplicateUniqueIndexError struct {
	index *StructIndex
	cause error
}

func (d DuplicateUniqueIndexError) Index() *StructIndex {
	return d.index
}

func (d DuplicateUniqueIndexError) Cause() error {
	return d.cause
}

func (d DuplicateUniqueIndexError) Error() string {
	return "duplicate unique index of " + d.index.Model.Type.PkgPath() +
		"." + d.index.Model.Type.Name() + " " + fmt.Sprint(d.index.FieldsNames()) +
		" caused by: " + d.cause.Error()
}

func IsDuplicateUniqueIndexError(err ...error) bool {
	return ErrorByType(reflect.TypeOf(DuplicateUniqueIndexError{}), err...) != nil
}

func GetDuplicateUniqueIndexError(err ...error) *DuplicateUniqueIndexError {
	if result := ErrorByType(reflect.TypeOf(DuplicateUniqueIndexError{}), err...); result != nil {
		return result.(*DuplicateUniqueIndexError)
	}
	return nil
}

type PathError interface {
	error
	Path() string
}
