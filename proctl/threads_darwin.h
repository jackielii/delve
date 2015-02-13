#include <errno.h>
#include <stdio.h>
#include <sys/types.h>
#include <mach/mach.h>
#include <mach/mach_vm.h>

int
acquire_mach_task(int, mach_port_name_t *);

int
write_memory(mach_port_name_t, mach_vm_address_t, void *, mach_msg_type_number_t);

int
read_memory(mach_port_name_t, mach_vm_address_t, void *, mach_msg_type_number_t);

x86_thread_state64_t
get_registers(mach_port_name_t);

void
set_pc(thread_act_t task, uint64_t pc);
