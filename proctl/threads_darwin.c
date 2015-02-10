#include "threads_darwin.h"

int
acquire_mach_task(int tid, mach_port_name_t *task) {
	kern_return_t kret;
	kret = task_for_pid(mach_task_self(), tid, task);
	if (kret != KERN_SUCCESS) return -1;
	return 0;
}

int
write_memory(mach_port_name_t task, mach_vm_address_t addr, void *d, mach_msg_type_number_t len) {
	kern_return_t kret;
	pointer_t data;
	memcpy((void *)&data, d, len);

	// Set permissions to enable writting to this memory location
	kret = mach_vm_protect(task, addr, len, FALSE, VM_PROT_READ | VM_PROT_WRITE | VM_PROT_COPY);
	if (kret != KERN_SUCCESS) return -1;

	kret = mach_vm_write((vm_map_t)task, addr, (vm_offset_t)&data, len);
	if (kret != KERN_SUCCESS) return -1;

	// Restore virtual memory permissions
	// TODO(dp) this should take into account original permissions somehow
	kret = mach_vm_protect(task, addr, len, FALSE, VM_PROT_READ | VM_PROT_EXECUTE);
	if (kret != KERN_SUCCESS) return -1;

	return 0;
}

int
read_memory(mach_port_name_t task, mach_vm_address_t addr, void *d, mach_msg_type_number_t len) {
	kern_return_t kret;
	pointer_t data;
	mach_msg_type_number_t count;

	kret = mach_vm_read((vm_map_t)task, addr, len, &data, &count);
	if (kret != KERN_SUCCESS) return -1;
	memcpy(d, (void *)data, len);
	return count;
}

x86_thread_state64_t
get_registers(mach_port_name_t task) {
	x86_thread_state64_t state;
	mach_msg_type_number_t stateCount = x86_THREAD_STATE64_COUNT;

	thread_get_state(task, x86_THREAD_STATE64, (thread_state_t)&state, &stateCount);
	return state;
}
