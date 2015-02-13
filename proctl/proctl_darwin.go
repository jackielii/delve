package proctl

/*
#include <stdio.h>
#include <libproc.h>
#include <sys/types.h>
#include <mach/mach.h>

static const unsigned char info_plist[]
__attribute__ ((section ("__TEXT,__info_plist"),used)) =
  "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
  "<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\""
  " \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n"
  "<plist version=\"1.0\">\n"
  "<dict>\n"
  "  <key>CFBundleIdentifier</key>\n"
  "  <string>org.dlv</string>\n"
  "  <key>CFBundleName</key>\n"
  "  <string>delve</string>\n"
  "  <key>CFBundleVersion</key>\n"
  "  <string>1.0</string>\n"
  "  <key>SecTaskAccess</key>\n"
  "  <array>\n"
  "    <string>allowed</string>\n"
  "    <string>debug</string>\n"
  "  </array>\n"
  "</dict>\n"
  "</plist>\n";

char *
findExecutable(int pid) {
	static char pathbuf[PATH_MAX];
	proc_pidpath(pid, pathbuf, PATH_MAX);
	return pathbuf;
}

int
get_task_threads(int tid, task_t task, thread_act_t *thread) {
	int pid;
	kern_return_t kret;
	thread_act_array_t list;
	mach_msg_type_number_t count;

	kret = task_threads(task, &list, &count);
	printf("count %d\n", count);
	if (kret != KERN_SUCCESS) return -1;
	for (int i = 0; i < (int)count; i++) {
		pid_for_task((mach_port_name_t)task, &pid);
		if (pid == tid) {
			*thread = list[i];
			return 0;
		}
	}
	return -1;
}
*/
import "C"
import (
	"debug/gosym"
	"debug/macho"
	"fmt"
	"os"
	"sync"

	"github.com/derekparker/delve/dwarf/frame"
	sys "golang.org/x/sys/unix"
)

func (dbp *DebuggedProcess) addThread(tid int) (*ThreadContext, error) {
	thread := &ThreadContext{
		Id:      tid,
		Process: dbp,
		os:      new(OSSpecificDetails),
	}
	dbp.Threads[tid] = thread
	if err := acquireMachTask(thread); err != nil {
		return nil, err
	}
	// TODO(dp) figure out a way to extract the correct thread_t for this thread
	var th C.thread_act_t
	ret := C.get_task_threads(C.int(tid), C.task_t(thread.os.task), &th)
	if ret < 0 {
		return nil, fmt.Errorf("could not get mach thread for %d", tid)
	}
	thread.os.thread_act = th
	return thread, nil
}

// Finds the executable and then uses it
// to parse the following information:
// * Dwarf .debug_frame section
// * Dwarf .debug_line section
// * Go symbol table.
func (dbp *DebuggedProcess) LoadInformation() error {
	var (
		wg  sync.WaitGroup
		exe *macho.File
		err error
	)

	exe, err = dbp.findExecutable()
	if err != nil {
		return err
	}
	data, err := exe.DWARF()
	if err != nil {
		return err
	}
	dbp.Dwarf = data

	wg.Add(2)
	go dbp.parseDebugFrame(exe, &wg)
	go dbp.obtainGoSymbols(exe, &wg)
	wg.Wait()

	return nil
}

func (dbp *DebuggedProcess) parseDebugFrame(exe *macho.File, wg *sync.WaitGroup) {
	defer wg.Done()

	if sec := exe.Section("__debug_frame"); sec != nil {
		debugFrame, err := exe.Section("__debug_frame").Data()
		if err != nil {
			fmt.Println("could not get __debug_frame section", err)
			os.Exit(1)
		}
		dbp.FrameEntries = frame.Parse(debugFrame)
	} else {
		fmt.Println("could not find __debug_frame section in binary")
		os.Exit(1)
	}
}

func (dbp *DebuggedProcess) obtainGoSymbols(exe *macho.File, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		symdat  []byte
		pclndat []byte
		err     error
	)

	if sec := exe.Section("__gosymtab"); sec != nil {
		symdat, err = sec.Data()
		if err != nil {
			fmt.Println("could not get .gosymtab section", err)
			os.Exit(1)
		}
	}

	if sec := exe.Section("__gopclntab"); sec != nil {
		pclndat, err = sec.Data()
		if err != nil {
			fmt.Println("could not get .gopclntab section", err)
			os.Exit(1)
		}
	}

	pcln := gosym.NewLineTable(pclndat, exe.Section("__text").Addr)
	tab, err := gosym.NewTable(symdat, pcln)
	if err != nil {
		fmt.Println("could not get initialize line table", err)
		os.Exit(1)
	}

	dbp.GoSymTable = tab
}

// TODO(darwin) IMPLEMENT ME
func stopped(pid int) bool {
	return false
}

func (dbp *DebuggedProcess) findExecutable() (*macho.File, error) {
	pathptr, err := C.findExecutable(C.int(dbp.Pid))
	if err != nil {
		return nil, err
	}
	path := C.GoString(pathptr)
	return macho.Open(path)
}

func trapWait(dbp *DebuggedProcess, pid int) (int, *sys.WaitStatus, error) {
	var (
		wpid   int
		status *sys.WaitStatus
		err    error
	)

	for {
		wpid, status, err = wait(pid, 0)
		if err != nil {
			return -1, nil, fmt.Errorf("wait err %s %d", err, pid)
		}
		if wpid != 0 {
			break
		}
	}

	if th, ok := dbp.Threads[wpid]; ok {
		th.Status = status
	}
	if status.Exited() && wpid == dbp.Pid {
		return -1, status, ProcessExitedError{wpid}
	}
	if status.StopSignal() == sys.SIGTRAP {
		fmt.Println(status.StopSignal())
		return wpid, status, nil
	}
	return -1, nil, fmt.Errorf("wait: %s", status.StopSignal())
}

func wait(pid, options int) (int, *sys.WaitStatus, error) {
	var status sys.WaitStatus
	wpid, err := sys.Wait4(pid, &status, options, nil)
	return wpid, &status, err
}
