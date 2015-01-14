package proctl

import "C"
import "fmt"

type OSSpecificDetails struct {
	task C.task_t
}

// TODO(darwin)
func (t *ThreadContext) Halt() error {
	return fmt.Errorf("not implemented")
}

// TODO(darwin)
func writeMemory(tid int, addr uintptr, data []byte) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

// TODO(darwin)
func readMemory(tid int, addr uintptr, data []byte) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

// TODO(darwin)
func registers(tid int) (Registers, error) {
	return nil, fmt.Errorf("not implemented")
}
