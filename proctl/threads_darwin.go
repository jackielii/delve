package proctl

/*
#include <errno.h>
#include <stdio.h>
#include <sys/types.h>
#include <mach/mach.h>
#include <mach/mach_vm.h>

// page size
static vm_size_t mach_page_size;

int
acquire_mach_task(int tid, mach_port_name_t *task) {
	kern_return_t kret;
	kret = task_for_pid(mach_task_self(), tid, task);
	if (kret != KERN_SUCCESS) return -1;
	return 0;
}

int
write_memory(mach_port_name_t *task, mach_vm_address_t addr, pointer_t data, mach_msg_type_number_t len) {
	kern_return_t kret;
	kret = mach_vm_write((vm_map_t)*task, addr, data, len);
	if (kret != KERN_SUCCESS) {
		printf("%s\n", mach_error_string(kret));
		return -1;
	}
	return 0;
}

int
read_memory(mach_port_name_t *task, mach_vm_address_t addr, void *data, mach_msg_type_number_t len) {
	kern_return_t kret;
	mach_msg_type_number_t count;
	kret = mach_vm_read((vm_map_t)*task, addr, len, (vm_offset_t *)data, &count);
	if (kret != KERN_SUCCESS) return -1;
	return count;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type OSSpecificDetails struct {
	task C.mach_port_name_t
}

func acquireMachTask(thread *ThreadContext) error {
	if ret := C.acquire_mach_task(C.int(thread.Id), &thread.os.task); ret < 0 {
		return fmt.Errorf("could not acquire mach task %d", ret)
	}
	return nil
}

// TODO(darwin)
func (t *ThreadContext) Halt() error {
	return fmt.Errorf("halt not implemented")
}

func writeMemory(thread *ThreadContext, addr uintptr, data []byte) (int, error) {
	fmt.Printf("%#v\n", addr)
	vm_data := C.pointer_t(uintptr(unsafe.Pointer(&data[0])))
	ret := C.write_memory(&thread.os.task, C.mach_vm_address_t(addr), vm_data, C.mach_msg_type_number_t(len(data)))
	if ret < 0 {
		return 0, fmt.Errorf("could not write memory")
	}
	return len(data), nil
}

func readMemory(thread *ThreadContext, addr uintptr, data []byte) (int, error) {
	vm_data := unsafe.Pointer(&data[0])
	ret := C.read_memory(&thread.os.task, C.mach_vm_address_t(addr), vm_data, C.mach_msg_type_number_t(len(data)))
	if ret < 0 {
		return 0, fmt.Errorf("could not read memory")
	}
	return len(data), nil
}

// TODO(darwin)
func registers(tid int) (Registers, error) {
	return nil, fmt.Errorf("registers not implemented")
}
