package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/solo-io/solo-kit/pkg/utils/syncutils"

	"github.com/solo-io/go-utils/log"

	"io/ioutil"

	"time"

	"io"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"
)

const defaultVaultDockerImage = "vault:1.1.3"

type VaultFactory struct {
	vaultpath string
	tmpdir    string
	Port      int
}

func NewVaultFactory() (*VaultFactory, error) {
	vaultpath := os.Getenv("VAULT_BINARY")

	if vaultpath == "" {
		vaultPath, err := exec.LookPath("vault")
		if err == nil {
			log.Printf("Using vault from PATH: %s", vaultPath)
			vaultpath = vaultPath
		}
	}

	port := AllocateParallelPort(8200)

	if vaultpath != "" {
		return &VaultFactory{
			vaultpath: vaultpath,
			Port:      port,
		}, nil
	}

	// try to grab one form docker...
	tmpdir, err := ioutil.TempDir(os.Getenv("HELPER_TMP"), "vault")
	if err != nil {
		return nil, err
	}

	bash := fmt.Sprintf(`
set -ex
CID=$(docker run -d  %s /bin/sh -c exit)

# just print the image sha for repoducibility
echo "Using Vault Image:"
docker inspect %s -f "{{.RepoDigests}}"

docker cp $CID:/bin/vault .
docker rm -f $CID
    `, defaultVaultDockerImage, defaultVaultDockerImage)
	scriptfile := filepath.Join(tmpdir, "getvault.sh")

	ioutil.WriteFile(scriptfile, []byte(bash), 0755)

	cmd := exec.Command("bash", scriptfile)
	cmd.Dir = tmpdir
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &VaultFactory{
		vaultpath: filepath.Join(tmpdir, "vault"),
		tmpdir:    tmpdir,
		Port:      port,
	}, nil
}

func (ef *VaultFactory) Clean() error {
	if ef == nil {
		return nil
	}
	if ef.tmpdir != "" {
		os.RemoveAll(ef.tmpdir)

	}
	return nil
}

type VaultInstance struct {
	vaultpath string
	tmpdir    string
	cmd       *exec.Cmd
	token     string
	Port      int
}

func (ef *VaultFactory) NewVaultInstance() (*VaultInstance, error) {
	// try to grab one form docker...
	tmpdir, err := ioutil.TempDir(os.Getenv("HELPER_TMP"), "vault")
	if err != nil {
		return nil, err
	}

	return &VaultInstance{
		vaultpath: ef.vaultpath,
		tmpdir:    tmpdir,
		Port:      ef.Port,
	}, nil

}

func (i *VaultInstance) Run() error {
	return i.RunWithPort()
}

func (i *VaultInstance) Token() string {
	return i.token
}

func (i *VaultInstance) RunWithPort() error {
	cmd := exec.Command(i.vaultpath,
		"server",
		"-dev",
		"-dev-root-token-id=root",
		fmt.Sprintf("-dev-listen-address=0.0.0.0:%v", i.Port),
	)
	buf := &syncutils.Buffer{}
	w := io.MultiWriter(ginkgo.GinkgoWriter, buf)
	cmd.Dir = i.tmpdir
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Start()
	if err != nil {
		return err
	}
	i.cmd = cmd
	time.Sleep(time.Millisecond * 2500)

	tokenSlice := regexp.MustCompile("Root Token: ([\\-[:word:]]+)").FindAllString(buf.String(), 1)
	if len(tokenSlice) < 1 {
		return errors.Errorf("%s did not contain root token", buf.String())
	}

	i.token = strings.TrimPrefix(tokenSlice[0], "Root Token: ")

	enableCmd := exec.Command(i.vaultpath,
		"secrets",
		"enable",
		fmt.Sprintf("-address=http://127.0.0.1:%v", i.Port),
		"-version=2",
		"kv")
	enableCmd.Env = append(enableCmd.Env, "VAULT_TOKEN="+i.token)

	// enable kv storage
	enableCmdOut, err := enableCmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "enabling kv storage failed: %s", enableCmdOut)
	}
	return nil
}

func (i *VaultInstance) Binary() string {
	return i.vaultpath
}

func (i *VaultInstance) Clean() error {
	if i.cmd != nil {
		i.cmd.Process.Kill()
		i.cmd.Wait()
	}
	if i.tmpdir != "" {
		os.RemoveAll(i.tmpdir)
	}
	return nil
}

func (i *VaultInstance) Exec(args ...string) (string, error) {
	cmd := exec.Command(i.vaultpath, args...)
	cmd.Env = os.Environ()
	// disable DEBUG=1 from getting through to nomad
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s (%v)", out, err)
	}
	return string(out), err
}
