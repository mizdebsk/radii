package rhsm

import (
	"fmt"
	"os"
	"strings"

	"github.com/mizdebsk/radii/internal/api"
	"github.com/mizdebsk/radii/internal/log"
	"github.com/mizdebsk/radii/internal/sysinfo"
)

const (
	defaultRedhatRepoPath = "/etc/yum.repos.d/redhat.repo"
	defaultRhsmExecPath   = "/usr/sbin/subscription-manager"
)

type repoMgr struct {
	rhsmEnabled    bool
	systemInfo     sysinfo.SysInfo
	executor       api.Executor
	redhatRepoPath string
	rhsmExecPath   string
}

var _ api.RepositoryManager = (*repoMgr)(nil)

func NewRepositoryManager(executor api.Executor, systemInfo sysinfo.SysInfo) api.RepositoryManager {
	return &repoMgr{
		rhsmEnabled:    true,
		systemInfo:     systemInfo,
		executor:       executor,
		redhatRepoPath: defaultRedhatRepoPath,
		rhsmExecPath:   defaultRhsmExecPath,
	}
}

func (rm *repoMgr) SetSubscriptionsEnabled(enabled bool) {
	rm.rhsmEnabled = enabled
}

func (rm *repoMgr) repoIDForChannel(channel string) string {
	return fmt.Sprintf("rhel-%d-for-%s-%s-rpms", rm.systemInfo.OsVersion, rm.systemInfo.Arch, strings.ToLower(channel))
}

func (rm *repoMgr) GetRepoIDs(channels []string) ([]string, error) {
	if !rm.rhsmEnabled || !rm.systemInfo.IsRhel || len(channels) == 0 {
		return nil, nil
	}
	states, err := readRepoStates(rm.redhatRepoPath)
	if err != nil {
		return nil, err
	}
	var repos, missing []string
	for _, channel := range channels {
		repo := rm.repoIDForChannel(channel)
		repos = append(repos, repo)
		if _, defined := states[repo]; !defined {
			missing = append(missing, repo)
		}
	}
	if len(missing) > 0 {
		err := fmt.Errorf("required RHEL repository definitions are missing from %s: %s", rm.redhatRepoPath, strings.Join(missing, ", "))
		return nil, fmt.Errorf("%w\nAlternatively, use --skip-subscriptions before the command to use already-configured DNF repositories.", rm.withSubscriptionHelp(err))
	}
	return repos, nil
}

func (rm *repoMgr) EnsureRepositoriesEnabled(channels []string) error {
	if len(channels) == 0 {
		return nil
	}
	if !rm.rhsmEnabled {
		log.Warnf("Skipping Red Hat Subscription Manager (RHSM) setup.")
		log.Warnf("Repositories must already be configured; packages will be installed from existing DNF sources.")
		return nil
	}
	if rm.systemInfo.IsRhel {
		log.Logf("detected RHEL %d", rm.systemInfo.OsVersion)
		if rm.subscriptionManagerPresent() {
			log.Logf("Subscription Manager is present")
			return rm.ensureChannelsEnabled(channels)
		} else {
			log.Warnf("Subscription Manager is absent.")
			log.Warnf("You may need to enable appropriate repositories yourself.")
		}
	} else {
		log.Warnf("This system is not RHEL.")
		log.Warnf("You may need to enable appropriate repositories yourself.")
	}
	return nil
}

func (rm *repoMgr) subscriptionManagerPresent() bool {
	stat, err := os.Stat(rm.rhsmExecPath)
	if err != nil || stat == nil {
		log.Debugf("stat %s failed: %v", rm.rhsmExecPath, err)
		return false
	}
	log.Debugf("stat %s: isRegular=%v mode=0%o", rm.rhsmExecPath, stat.Mode().IsRegular(), stat.Mode().Perm())
	return stat.Mode().IsRegular() && stat.Mode().Perm()&0111 != 0
}

func (rm *repoMgr) ensureChannelsEnabled(channels []string) error {
	log.Logf("checking repository status")
	states, err := readRepoStates(rm.redhatRepoPath)
	if err != nil {
		return err
	}
	allEnabled := true
	args := []string{"repos"}
	for _, channel := range channels {
		repo := rm.repoIDForChannel(channel)
		log.Logf("mapped RHEL channel %s to repo ID %s", channel, repo)
		if !states[repo] {
			log.Infof("enabling channel %s, repository %s", channel, repo)
			args = append(args, "--enable", repo)
			allEnabled = false
		} else {
			log.Logf("repository %s is already enabled", repo)
		}
	}

	if allEnabled {
		log.Logf("all required repositories are already enabled")
		return nil
	}

	log.Logf("running subscription-manager to enable repositories")
	err = rm.executor.Run(rm.rhsmExecPath, args)
	if err != nil {
		return getDetailedSubscriptionError(err, rm.systemInfo)
	}

	log.Logf("repositories were enabled successfully")
	return nil
}

func (rm *repoMgr) withSubscriptionHelp(err error) error {
	return fmt.Errorf("%w\nPlease ensure the system is registered and subscribed (see https://access.redhat.com/solutions/253273).", err)
}
