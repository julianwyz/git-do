package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/julianwyz/git-do/internal/cli"
	"github.com/julianwyz/git-do/internal/config"
	"github.com/julianwyz/git-do/internal/credentials"
)

type (
	testDst struct {
		wbuf bytes.Buffer
		rbuf bytes.Buffer
	}
	testFileInfo struct{}
	roundtrip    struct{}
)

func TestNew(t *testing.T) {
	os.Args = []string{
		"git-do",
		"help",
	}
	prog, err := cli.New()
	if err != nil {
		t.Fatal(err)
	}
	if prog == nil {
		t.Fatal("cli should not be nil")
	}
}

func TestCmd__Help(t *testing.T) {
	dir := setup(t)
	out := &testDst{}

	os.Args = []string{
		"git-do",
		"help",
	}
	prog, err := cli.New(
		cli.WithWorkingDir(dir),
		cli.WithHomeDir(dir),
		cli.WithInput(&testDst{}),
		cli.WithOutput(out),
	)
	if err != nil {
		t.Fatal(err)
	}
	if prog == nil {
		t.Fatal("cli should not be nil")
	}

	if err := prog.Exec(t.Context()); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(
		out.wbuf.String(),
		"git do commit",
	) {
		t.Fatal("unexpected help output")
	}
}

func TestCmd__Init__Fresh(t *testing.T) {
	dir, err := os.MkdirTemp("", "gitdo-test-init-*")
	if err != nil {
		t.Fatal(err)
	}

	out := &testDst{}

	os.Args = []string{
		"git-do",
		"init",
	}
	prog, err := cli.New(
		cli.WithWorkingDir(dir),
		cli.WithHomeDir(dir),
		cli.WithInput(&testDst{}),
		cli.WithOutput(out),
	)
	if err != nil {
		t.Fatal(err)
	}
	if prog == nil {
		t.Fatal("cli should not be nil")
	}

	if err := prog.Exec(t.Context()); err != nil {
		t.Fatal(err)
	}

	txt := out.wbuf.String()

	if !strings.Contains(
		txt,
		"Established git repo.",
	) {
		t.Fatal("expected git repo to be established")
	}

	if !strings.Contains(
		txt,
		"Created initial project configuration.",
	) {
		t.Fatal("expected project config to be made")
	}

	if !strings.Contains(
		txt,
		"Device is non-interactive. Using empty key in credentials.",
	) {
		t.Fatal("expected credentials file to be made")
	}
}

func TestCmd__Init__Existing(t *testing.T) {
	dir := setup(t)
	gitInit(t, dir)

	out := &testDst{}

	os.Args = []string{
		"git-do",
		"init",
	}
	prog, err := cli.New(
		cli.WithWorkingDir(dir),
		cli.WithHomeDir(dir),
		cli.WithInput(&testDst{}),
		cli.WithOutput(out),
	)
	if err != nil {
		t.Fatal(err)
	}
	if prog == nil {
		t.Fatal("cli should not be nil")
	}

	if err := prog.Exec(t.Context()); err != nil {
		t.Fatal(err)
	}

	txt := out.wbuf.String()
	if !strings.Contains(
		txt,
		"git repo already established.",
	) {
		t.Fatal("expected git repo to not be established")
	}

	if !strings.Contains(
		txt,
		"configuration file already exists.",
	) {
		t.Fatal("expected project config to not be made")
	}

	if !strings.Contains(
		txt,
		"Credentials file already exists.",
	) {
		t.Fatal("expected credentials file to not be made")
	}
}

func TestCmd__Commit(t *testing.T) {
	dir := setup(t)
	gitInit(t, dir)
	addFile(t, dir)
	out := &testDst{}

	os.Args = []string{
		"git-do",
		"commit",
	}
	prog, err := cli.New(
		cli.WithWorkingDir(dir),
		cli.WithHomeDir(dir),
		cli.WithInput(&testDst{}),
		cli.WithOutput(out),
		cli.WithHTTPClient(makeClient()),
	)
	if err != nil {
		t.Fatal(err)
	}
	if prog == nil {
		t.Fatal("cli should not be nil")
	}

	if err := prog.Exec(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func setup(t *testing.T) string {
	dir, err := os.MkdirTemp("", "gitdo-test-*")
	if err != nil {
		t.Fatal(err)
	}

	err = config.WriteDefault(
		dir,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = credentials.WriteDefault(dir, "test")
	if err != nil {
		t.Fatal(err)
	}

	return dir
}

func gitInit(t *testing.T, wd string) {
	cmd := exec.Command(
		"git", "init",
	)
	cmd.Dir = wd

	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func addFile(t *testing.T, wd string) {
	f, err := os.CreateTemp(wd, "file-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, _ = fmt.Fprintf(
		f,
		"%d", time.Now().UnixNano(),
	)

	cmd := exec.Command(
		"git", "add", f.Name(),
	)
	cmd.Dir = wd

	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func (recv *testDst) Write(p []byte) (n int, err error) {
	return recv.wbuf.Write(p)
}

func (recv *testDst) WriteString(s string) (n int, err error) {
	return recv.wbuf.WriteString(s)
}

func (recv *testDst) Stat() (os.FileInfo, error) {
	return &testFileInfo{}, nil
}

func (*testDst) Close() error {
	return nil
}

func (recv *testDst) Read(p []byte) (n int, err error) {
	return recv.rbuf.Read(p)
}

func (*testFileInfo) Name() string {
	return "foo-bar"
}

func (*testFileInfo) Size() int64 {
	return 123
}

func (*testFileInfo) Mode() os.FileMode {
	return 0
}

func (*testFileInfo) ModTime() time.Time {
	return time.Now()
}

func (*testFileInfo) IsDir() bool {
	return false
}
func (*testFileInfo) Sys() any {
	return nil
}

func makeClient() *http.Client {
	return &http.Client{
		Transport: &roundtrip{},
	}
}

func (recv *roundtrip) RoundTrip(req *http.Request) (res *http.Response, err error) {
	hdr := make(http.Header)
	hdr.Set("content-type", "application/json")

	res = &http.Response{
		Header: hdr,
		Body:   recv.makeBody(map[string]any{}),
	}

	return
}

func (recv *roundtrip) makeBody(obj any) io.ReadCloser {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	return io.NopCloser(
		bytes.NewBuffer(data),
	)
}
