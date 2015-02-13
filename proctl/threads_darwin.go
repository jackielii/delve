package proctl

// #include "threads_darwin.h"
import "C"
import (
	"fmt"
	"unsafe"
)

type OSSpecificDetails struct {
	task       C.mach_port_name_t
	thread_act C.thread_act_t
}

func acquireMachTask(thread *ThreadContext) error {
	if ret := C.acquire_mach_task(C.int(thread.Id), &thread.os.task); ret < 0 {
		return fmt.Errorf("could not acquire mach task %d", ret)
	}
	return nil
}

// TODO(darwin)
func (t *ThreadContext) Halt() error {
	var kret C.kern_return_t
	kret = C.task_suspend(C.task_t(t.os.task))
	if kret != C.KERN_SUCCESS {
		return fmt.Errorf("could not suspend thread %d", t.Id)
	}
	return nil
}

func writeMemory(thread *ThreadContext, addr uintptr, data []byte) (int, error) {
	var (
		vm_data = unsafe.Pointer(&data[0])
		vm_addr = C.mach_vm_address_t(addr)
		length  = C.mach_msg_type_number_t(len(data))
	)

	if ret := C.write_memory(thread.os.task, vm_addr, vm_data, length); ret < 0 {
		return 0, fmt.Errorf("could not write memory")
	}
	return len(data), nil
}

func readMemory(thread *ThreadContext, addr uintptr, data []byte) (int, error) {
	var (
		vm_data = unsafe.Pointer(&data[0])
		vm_addr = C.mach_vm_address_t(addr)
		length  = C.mach_msg_type_number_t(len(data))
	)

	ret := C.read_memory(thread.os.task, vm_addr, vm_data, length)
	if ret < 0 {
		return 0, fmt.Errorf("could not read memory")
	}
	return len(data), nil
}
