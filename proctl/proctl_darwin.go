package proctl

/*
#include <libproc.h>

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

// TODO(darwin)
func trapWait(dbp *DebuggedProcess, pid int) (int, *sys.WaitStatus, error) {
	return 0, nil, nil
}

func wait(pid, options int) (int, *sys.WaitStatus, error) {
	var status sys.WaitStatus
	wpid, err := sys.Wait4(pid, &status, options, nil)
	return wpid, &status, err
}
