package proctl

// #include "threads_darwin.h"
import "C"
import "fmt"

type Regs struct {
	pc, sp uint64
}

func (r *Regs) PC() uint64 {
	return r.pc
}

func (r *Regs) SP() uint64 {
	return r.sp
}

func (r *Regs) SetPC(tid int, pc uint64) error {
	return fmt.Errorf("setpc not implemented")
}

func registers(thread *ThreadContext) (Registers, error) {
	state := C.get_registers(thread.os.task)
	regs := &Regs{pc: uint64(state.__rip), sp: uint64(state.__rsp)}
	return regs, nil
}
