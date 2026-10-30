package rhsm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/log"
)

var partnerRepoDir = "/etc/yum.repos.d"

func partnerRepoFilePath(providerID string) string {
	return fmt.Sprintf("%s/radii-%s-partner.repo", partnerRepoDir, providerID)
}

func (rm *repoMgr) partnerReposOnDisk(repos []api.PartnerRepo) bool {
	needed := make(map[string]bool)
	for _, repo := range repos {
		needed[rm.resolveDNFVars(repo.RepoID)] = true
	}
	files, _ := filepath.Glob(partnerRepoDir + "/radii-*-partner.repo")
	for _, f := range files {
		sections, err := parseRepoFile(f)
		if err != nil {
			continue
		}
		for name := range sections {
			delete(needed, name)
		}
	}
	return len(needed) == 0
}

func (rm *repoMgr) resolveDNFVars(s string) string {
	s = strings.ReplaceAll(s, "$basearch", rm.systemInfo.Arch)
	s = strings.ReplaceAll(s, "$releasever_major", fmt.Sprintf("%d", rm.systemInfo.OsVersion))
	return s
}

func (rm *repoMgr) ensurePartnerRepos(providerID string, repos []api.PartnerRepo) error {
	certs, err := readSSLCertPaths(rm.redhatRepoPath, rm.systemInfo.OsVersion, rm.systemInfo.Arch)
	if err != nil {
		return fmt.Errorf("failed to read SSL certificates for partner repository: %w", err)
	}

	var buf strings.Builder
	for _, repo := range repos {
		fmt.Fprintf(&buf, "[%s]\n", rm.resolveDNFVars(repo.RepoID))
		fmt.Fprintf(&buf, "name=%s\n", rm.resolveDNFVars(repo.Name))
		fmt.Fprintf(&buf, "baseurl=%s\n", rm.resolveDNFVars(repo.BaseURL))
		fmt.Fprintf(&buf, "enabled=1\n")
		fmt.Fprintf(&buf, "gpgcheck=1\n")
		fmt.Fprintf(&buf, "gpgkey=%s\n", rm.resolveDNFVars(repo.GPGKey))
		fmt.Fprintf(&buf, "sslverify=1\n")
		fmt.Fprintf(&buf, "sslcacert=/etc/pki/tls/certs/ca-bundle.crt\n")
		fmt.Fprintf(&buf, "sslclientkey=%s\n", certs.clientKey)
		fmt.Fprintf(&buf, "sslclientcert=%s\n", certs.clientCert)
		fmt.Fprintf(&buf, "metadata_expire=86400\n")
		buf.WriteString("\n")
	}

	path := partnerRepoFilePath(providerID)
	log.Logf("writing partner repository configuration to %s", path)
	if err := os.WriteFile(path, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write partner repository file %s: %w", path, err)
	}
	return nil
}

func (rm *repoMgr) partnerRepoDNFArgs(repos []api.PartnerRepo) ([]string, error) {
	certs, err := readSSLCertPaths(rm.redhatRepoPath, rm.systemInfo.OsVersion, rm.systemInfo.Arch)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSL certificates for partner repository: %w", err)
	}

	var args []string
	for _, repo := range repos {
		repoID := rm.resolveDNFVars(repo.RepoID)
		args = append(args, fmt.Sprintf("--repofrompath=%s,%s", repoID, rm.resolveDNFVars(repo.BaseURL)))
		args = append(args, fmt.Sprintf("--setopt=%s.name=%s", repoID, rm.resolveDNFVars(repo.Name)))
		args = append(args, fmt.Sprintf("--setopt=%s.gpgcheck=1", repoID))
		args = append(args, fmt.Sprintf("--setopt=%s.gpgkey=%s", repoID, rm.resolveDNFVars(repo.GPGKey)))
		args = append(args, fmt.Sprintf("--setopt=%s.sslverify=1", repoID))
		args = append(args, fmt.Sprintf("--setopt=%s.sslcacert=/etc/pki/tls/certs/ca-bundle.crt", repoID))
		args = append(args, fmt.Sprintf("--setopt=%s.sslclientkey=%s", repoID, certs.clientKey))
		args = append(args, fmt.Sprintf("--setopt=%s.sslclientcert=%s", repoID, certs.clientCert))
		args = append(args, "--enablerepo="+repoID)
	}
	return args, nil
}

type sslCertPaths struct {
	clientKey  string
	clientCert string
}

func readSSLCertPaths(path string, osVersion int, arch string) (sslCertPaths, error) {
	sections, err := parseRepoFile(path)
	if err != nil {
		return sslCertPaths{}, err
	}
	baseosID := fmt.Sprintf("rhel-%d-for-%s-baseos-rpms", osVersion, arch)
	if kv, ok := sections[baseosID]; ok {
		key := kv["sslclientkey"]
		cert := kv["sslclientcert"]
		if key != "" && cert != "" {
			return sslCertPaths{clientKey: key, clientCert: cert}, nil
		}
	}
	for _, kv := range sections {
		key := kv["sslclientkey"]
		cert := kv["sslclientcert"]
		if key != "" && cert != "" {
			return sslCertPaths{clientKey: key, clientCert: cert}, nil
		}
	}
	return sslCertPaths{}, fmt.Errorf("SSL client certificate paths not found in %s", path)
}
