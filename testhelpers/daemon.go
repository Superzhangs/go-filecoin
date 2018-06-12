package testhelpers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/filecoin-project/go-filecoin/config"
)

// Daemon is a daemon
type Daemon struct {
	CmdAddr    string
	SwarmAddr  string
	RepoDir    string
	Init       bool
	waitMining bool

	// The filecoin daemon process
	process *exec.Cmd

	lk     sync.Mutex
	Stdin  io.Writer
	Stdout io.Reader
	Stderr io.Reader
}

// NewDaemon makes a new daemon
func NewDaemon(options ...func(*Daemon)) (*Daemon, error) {
	// Ensure we have the actual binary
	filecoinBin, err := GetFilecoinBinary()
	if err != nil {
		return nil, err
	}

	//Ask the kernel for a port to avoid conflicts
	cmdPort, err := GetFreePort()
	if err != nil {
		return nil, err
	}
	swarmPort, err := GetFreePort()
	if err != nil {
		return nil, err
	}

	dir, err := ioutil.TempDir("", "go-fil-test")
	if err != nil {
		return nil, err
	}

	d := &Daemon{
		CmdAddr:   fmt.Sprintf(":%d", cmdPort),
		SwarmAddr: fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", swarmPort),
		RepoDir:   dir,
		Init:      true, // we want to init unless told otherwise
	}

	// configure TestDaemon options
	for _, option := range options {
		option(d)
	}

	repodirFlag := fmt.Sprintf("--repodir=%s", d.RepoDir)
	if d.Init {
		out, err := runInit(repodirFlag)
		if err != nil {
			d.Info(string(out))
			return nil, err
		}
	}

	// define filecoin daemon process
	d.process = exec.Command(filecoinBin, "daemon",
		fmt.Sprintf("--repodir=%s", d.RepoDir),
		fmt.Sprintf("--cmdapiaddr=%s", d.CmdAddr),
		fmt.Sprintf("--swarmlisten=%s", d.SwarmAddr),
	)

	// setup process pipes
	d.Stdout, err = d.process.StdoutPipe()
	if err != nil {
		return nil, err
	}
	d.Stderr, err = d.process.StderrPipe()
	if err != nil {
		return nil, err
	}
	d.Stdin, err = d.process.StdinPipe()
	if err != nil {
		return nil, err
	}

	return d, nil
}

func runInit(opts ...string) ([]byte, error) {
	return runCommand("init", opts...)
}

func runCommand(cmd string, opts ...string) ([]byte, error) {
	filecoinBin, err := GetFilecoinBinary()
	if err != nil {
		return nil, err
	}

	process := exec.Command(filecoinBin, append([]string{cmd}, opts...)...)
	return process.CombinedOutput()
}

// Logf is a daemon logger
// TODO print the daemon api like `Log` see below
func (d *Daemon) Logf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// Log is a daemon logger
func (d *Daemon) Info(msg ...string) {
	fmt.Printf("[%s]\t %s", d.CmdAddr, msg)
}

// Log is a daemon logger
func (d *Daemon) Error(err error) {
	fmt.Errorf("[%s]\t %s", d.CmdAddr, err)
}

// Run runs commands on the daemon
func (d *Daemon) Run(args ...string) (*Output, error) {
	return d.RunWithStdin(nil, args...)
}

// RunWithStdin runs things with stdin
func (d *Daemon) RunWithStdin(stdin io.Reader, args ...string) (*Output, error) {
	bin, err := GetFilecoinBinary()
	if err != nil {
		return nil, err
	}

	// handle Run("cmd subcmd")
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}

	finalArgs := append(args, "--repodir="+d.RepoDir, "--cmdapiaddr="+d.CmdAddr)

	d.Logf("run: %q", strings.Join(finalArgs, " "))
	cmd := exec.Command(bin, finalArgs...)

	if stdin != nil {
		cmd.Stdin = stdin
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	stderrBytes, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	stdoutBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	o := &Output{
		Args:   args,
		Stdout: stdout,
		stdout: stdoutBytes,
		Stderr: stderr,
		stderr: stderrBytes,
	}

	err = cmd.Wait()

	switch err := err.(type) {
	case *exec.ExitError:
		// TODO: its non-trivial to get the 'exit code' cross platform...
		o.Code = 1
	default:
		o.Error = err
	case nil:
		// okay
	}

	return o, nil
}

// ReadStdout reads that
func (d *Daemon) ReadStdout() string {
	d.lk.Lock()
	defer d.lk.Unlock()
	out, err := ioutil.ReadAll(d.Stdout)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// ReadStderr reads that
func (d *Daemon) ReadStderr() string {
	d.lk.Lock()
	defer d.lk.Unlock()
	out, err := ioutil.ReadAll(d.Stderr)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// Start starts the daemon process
func (d *Daemon) Start() (*Daemon, error) {
	if err := d.process.Start(); err != nil {
		return nil, err
	}
	if err := d.WaitForAPI(); err != nil {
		return nil, err
	}
	return d, nil
}

// Shutdown suts things down
// TODO don't panic be happy
func (d *Daemon) Shutdown() {
	if err := d.process.Process.Signal(syscall.SIGTERM); err != nil {
		d.Logf("Daemon Stderr:\n%s", d.ReadStderr())
		d.Logf("Failed to kill daemon %s", err)
		panic(err)
	}

	if d.RepoDir == "" {
		panic("testdaemon had no repodir set")
	}

	_ = os.RemoveAll(d.RepoDir)
}

// ShutdownSuccess needs comments
// TODO don't panic be happy
func (d *Daemon) ShutdownSuccess() {
	if err := d.process.Process.Signal(syscall.SIGTERM); err != nil {
		panic(err)
	}
	dOut := d.ReadStderr()
	if strings.Contains(dOut, "ERROR") {
		panic("Daemon has error messages")
	}
}

// ShutdownEasy needs comments
// TODO don't panic be happy
func (d *Daemon) ShutdownEasy() {
	if err := d.process.Process.Signal(syscall.SIGINT); err != nil {
		panic(err)
	}
	dOut := d.ReadStderr()
	if strings.Contains(dOut, "ERROR") {
		d.Info("Daemon has error messages")
	}
}

// WaitForAPI waits for the daemon to be running by hitting the http endpoint
func (d *Daemon) WaitForAPI() error {
	for i := 0; i < 100; i++ {
		err := TryAPICheck(d)
		if err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 100)
	}
	return fmt.Errorf("filecoin node failed to come online in given time period (20 seconds)")
}

// Config is a helper to read out the config of the deamon
// TODO don't panic be happy
func (d *Daemon) Config() *config.Config {
	cfg, err := config.ReadFile(filepath.Join(d.RepoDir, "config.toml"))
	if err != nil {
		panic(err)
	}
	return cfg
}

// MineAndPropagate mines a block and ensure the block has propogated to all `peers`
// by comparing the current head block of `d` with the head block of each peer in `peers`
// TODO don't panic be happy
func (d *Daemon) MineAndPropagate(wait time.Duration, peers ...*Daemon) {
	_, err := d.Run("mining", "once")
	if err != nil {
		panic(err)
	}
	// short circuit
	if peers == nil {
		return
	}
	// ensure all peers have same chain head as `d`
	d.MustHaveChainHeadBy(wait, peers)
}

// TryAPICheck will check if the daemon is ready
// TODO don't panic be happy
func TryAPICheck(d *Daemon) error {
	url := fmt.Sprintf("http://127.0.0.1%s/api/id", d.CmdAddr)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	out := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return fmt.Errorf("liveness check failed: %s", err)
	}

	_, ok := out["ID"]
	if !ok {
		return fmt.Errorf("liveness check failed: ID field not present in output")
	}

	return nil
}

func (d *Daemon) SetWaitMining(t bool) {
	d.lk.Lock()
	defer d.lk.Unlock()
	d.waitMining = t
}

func (d *Daemon) WaitMining() bool {
	d.lk.Lock()
	defer d.lk.Unlock()
	return d.waitMining
}
