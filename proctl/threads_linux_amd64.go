package proctl

import sys "golang.org/x/sys/unix"

// Not actually used, but necessary
// to be defined.
type OSSpecificDetails interface{}

func (t *ThreadContext) Halt() error {
	if stopped(t.Id) {
		return nil
	}
	return sys.Tgkill(t.Process.Pid, t.Id, sys.SIGSTOP)
}

func writeMemory(tid int, addr uintptr, data []byte) (int, error) {
	return sys.PtracePokeData(tid, addr, data)
}

func readMemory(tid int, addr uintptr, data []byte) (int, error) {
	return sys.PtracePeekData(tid, addr, data)
}
