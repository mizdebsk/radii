package rhsm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

var testPartnerRepos = []api.PartnerRepo{
	{
		RepoID:  "radii-partner-nvidia-cuda",
		Name:    "radii - NVIDIA CUDA Partner Repository",
		BaseURL: "https://packages.redhat.com/api/pulp-content/partner-repositories/@partner-repos-nvidia-rm/cuda-monorepo-rhel-$releasever_major/custom-1-$basearch/",
		GPGKey:  "https://redhat.gitlab.io/partner-repositories/docs/keys/nvidia-cuda-monorepo-rhel-$releasever_major.asc",
	},
}

func TestReadSSLCertPaths(t *testing.T) {
	certs, err := readSSLCertPaths("testdata/rhel10.repo", 10, "x86_64")
	if err != nil {
		t.Fatal(err)
	}
	if certs.clientKey != "/etc/pki/entitlement/13375007541689042-key.pem" {
		t.Errorf("clientKey = %q", certs.clientKey)
	}
	if certs.clientCert != "/etc/pki/entitlement/13375007541689042.pem" {
		t.Errorf("clientCert = %q", certs.clientCert)
	}
}

func TestReadSSLCertPathsMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.repo")
	if err := os.WriteFile(path, []byte("[test]\nenabled=1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := readSSLCertPaths(path, 10, "x86_64")
	if err == nil || !strings.Contains(err.Error(), "SSL client certificate") {
		t.Fatalf("expected SSL cert error, got %v", err)
	}
}

func TestReadSSLCertPathsFileNotFound(t *testing.T) {
	_, err := readSSLCertPaths(filepath.Join(t.TempDir(), "missing.repo"), 10, "x86_64")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestEnsurePartnerRepos(t *testing.T) {
	dir := t.TempDir()
	partnerRepoDir = dir
	defer func() { partnerRepoDir = "/etc/yum.repos.d" }()
	rm := repoMgr{
		systemInfo:     sysinfo.SysInfo{Arch: "x86_64", OsVersion: 10},
		redhatRepoPath: "testdata/rhel10.repo",
	}
	if err := rm.ensurePartnerRepos("nvidia", testPartnerRepos); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "radii-nvidia-partner.repo"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "[radii-partner-nvidia-cuda]") {
		t.Error("missing [radii-partner-nvidia-cuda] section")
	}
	if !strings.Contains(s, "radii - NVIDIA CUDA Partner Repository") {
		t.Error("missing descriptive repo name")
	}
	if !strings.Contains(s, "custom-1-x86_64") {
		t.Error("expected $basearch resolved to x86_64 in baseurl")
	}
	if !strings.Contains(s, "sslclientkey=/etc/pki/entitlement/13375007541689042-key.pem") {
		t.Error("sslclientkey not extracted from redhat.repo")
	}
	if !strings.Contains(s, "sslclientcert=/etc/pki/entitlement/13375007541689042.pem") {
		t.Error("sslclientcert not extracted from redhat.repo")
	}
	if !strings.Contains(s, "enabled=1") {
		t.Error("partner repo should be enabled")
	}
	if !strings.Contains(s, "gpgcheck=1") {
		t.Error("partner repo should have gpgcheck enabled")
	}
	if !strings.Contains(s, "cuda-monorepo-rhel-10") {
		t.Error("expected $releasever_major resolved to 10 in baseurl")
	}
	if !strings.Contains(s, "gpgkey=https://redhat.gitlab.io/partner-repositories/docs/keys/nvidia-cuda-monorepo-rhel-10.asc") {
		t.Error("partner repo should have GPG key URL with resolved $releasever_major")
	}
}

func TestEnsurePartnerReposMultiple(t *testing.T) {
	dir := t.TempDir()
	partnerRepoDir = dir
	defer func() { partnerRepoDir = "/etc/yum.repos.d" }()
	rm := repoMgr{
		systemInfo:     sysinfo.SysInfo{Arch: "x86_64", OsVersion: 10},
		redhatRepoPath: "testdata/rhel10.repo",
	}
	repos := []api.PartnerRepo{
		{RepoID: "repo-one", Name: "First Repo", BaseURL: "https://example.com/one/$basearch/", GPGKey: "https://example.com/gpg/one"},
		{RepoID: "repo-two", Name: "Second Repo", BaseURL: "https://example.com/two/$basearch/", GPGKey: "https://example.com/gpg/two"},
	}
	if err := rm.ensurePartnerRepos("testprov", repos); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "radii-testprov-partner.repo"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "[repo-one]") {
		t.Error("missing [repo-one] section")
	}
	if !strings.Contains(s, "[repo-two]") {
		t.Error("missing [repo-two] section")
	}
}

func TestPartnerRepoDNFArgs(t *testing.T) {
	rm := repoMgr{
		systemInfo:     sysinfo.SysInfo{Arch: "x86_64", OsVersion: 10},
		redhatRepoPath: "testdata/rhel10.repo",
	}
	args, err := rm.partnerRepoDNFArgs(testPartnerRepos)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--repofrompath=radii-partner-nvidia-cuda,") {
		t.Error("missing --repofrompath argument")
	}
	if !strings.Contains(joined, "--setopt=radii-partner-nvidia-cuda.gpgcheck=1") {
		t.Error("missing gpgcheck setopt")
	}
	if !strings.Contains(joined, "--setopt=radii-partner-nvidia-cuda.sslclientkey=/etc/pki/entitlement/13375007541689042-key.pem") {
		t.Error("missing sslclientkey setopt")
	}
	if !strings.Contains(joined, "--enablerepo=radii-partner-nvidia-cuda") {
		t.Error("missing --enablerepo argument")
	}
}

func TestGetRepoIDsNoPartner(t *testing.T) {
	rm := repoMgr{
		rhsmEnabled:    true,
		systemInfo:     sysinfo.SysInfo{IsRhel: true, OsVersion: 10, Arch: "x86_64"},
		redhatRepoPath: "testdata/rhel10.repo",
		rhsmExecPath:   "testdata/rhsm-exec",
	}
	repos, err := rm.GetRepoIDs([]string{"BaseOS"})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || repos[0] != "rhel-10-for-x86_64-baseos-rpms" {
		t.Fatalf("repos = %v, want [rhel-10-for-x86_64-baseos-rpms]", repos)
	}
}

func TestEnsurePartnerReposEnabledCreatesFile(t *testing.T) {
	dir := t.TempDir()
	partnerRepoDir = dir
	defer func() { partnerRepoDir = "/etc/yum.repos.d" }()
	rm := repoMgr{
		systemInfo:     sysinfo.SysInfo{Arch: "x86_64", OsVersion: 10},
		redhatRepoPath: "testdata/rhel10.repo",
	}
	if err := rm.EnsurePartnerReposEnabled("nvidia", testPartnerRepos); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "radii-nvidia-partner.repo")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("EnsurePartnerReposEnabled should create partner repo file: %v", err)
	}
}

func TestEnsurePartnerReposEnabledEmpty(t *testing.T) {
	rm := repoMgr{}
	if err := rm.EnsurePartnerReposEnabled("nvidia", nil); err != nil {
		t.Fatal(err)
	}
}
