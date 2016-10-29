package bouncermain

import (
	"reflect"
	"time"
)

// SetField assigns value v to field f of struct s.
func SetField(s interface{}, f string, v interface{}) error {
	sValue := reflect.ValueOf(s).Elem()
	sField := sValue.FieldByName(f)

	if !sField.IsValid() {
		return nil
	}

	if !sField.CanSet() {
		return nil
	}

	sFieldType := sField.Type()
	val := reflect.ValueOf(v).Convert(sFieldType)

	sField.Set(val)

	return nil
}

func RecvTimeout(ch <-chan bool, d time.Duration) (v bool, err error) {
	switch {

	case d < 0:
		// timeout negative, wait forever
		v = <-ch

	case d == 0:
		// timeout zero, return immediately
		select {
		case v = <-ch:
		default:
			err = ErrTimedOut
		}

	case d > 0:
		// timeout positive, wait until timeout
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case v = <-ch:
		case <-timer.C:
			err = ErrTimedOut
		}
	}
	return v, err
}
