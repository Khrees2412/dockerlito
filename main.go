package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("invalid command")
	}
}

func run() {
	fmt.Printf("Running %v as %v \n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		// UidMappings: ,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	cmd.Run()
}

func child() {
	fmt.Printf("running %v as %v \n", os.Args[2:], os.Getpid())
	cgroup()

	syscall.Sethostname([]byte("dockerlito"))
	syscall.Chroot("/mp/ubuntu-fs")
	syscall.Chdir("/")
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Run()
	syscall.Unmount("/proc", 0)
}

// docker run image <cmd> <params>
// go run main.go run  <cmd> <params>

// limit what a process can see
func cgroup() {
	cgroups := "/sys/fs/cgroup"
	cgroupPath := filepath.Join(cgroups, "chris")

	err := os.Mkdir(cgroupPath, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	// Enable the pids controller
	must(os.WriteFile(filepath.Join(cgroups, "cgroup.subtree_control"), []byte("+pids"), 0700))

	// Set the pids limit (equivalent to pids.max)
	must(os.WriteFile(filepath.Join(cgroupPath, "pids.max"), []byte("20"), 0700))

	// Add the current process to the cgroup (cgroup.procs replaces group.procs)
	must(os.WriteFile(filepath.Join(cgroupPath, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
