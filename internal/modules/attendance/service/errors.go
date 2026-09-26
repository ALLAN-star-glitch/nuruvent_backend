// internal/modules/attendance/service/errors.go

package service

import "fmt"



func errMissingDependency(name string) error {
	return fmt.Errorf("attendance service: missing dependency %s", name)
}

