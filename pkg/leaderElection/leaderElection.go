package leaderElection

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	v1 "k8s.io/client-go/kubernetes/typed/coordination/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// official k8s example: https://github.com/kubernetes/client-go/blob/master/examples/leader-election

type LeaderElection struct {
	k8sClientLeaseGetter *v1.LeasesGetter
	logger               logger.Logger
	isLeader             bool
}

func New(k8sClientLeaseGetter *v1.LeasesGetter, internalLogger logger.Logger) *LeaderElection {
	if k8sClientLeaseGetter == nil {
		internalLogger.LogAndReportUnexpectedError("leaderElection: k8s client is nil")
		return &LeaderElection{
			k8sClientLeaseGetter: nil,
			logger:               internalLogger,
			isLeader:             true,
		}
	} else {
		le := &LeaderElection{
			k8sClientLeaseGetter: k8sClientLeaseGetter,
			logger:               internalLogger,
			isLeader:             false,
		}
		// le.listenForChangesInLeader is a blocking function call, therefore we run it in a goroutine
		// we also wait for the first leader election to be done, before returning the leaderElection object, with a 5000ms timeout
		hasSucceededFirstLeaderElectionChannel := make(chan bool)
		go le.listenForChangesInLeader(hasSucceededFirstLeaderElectionChannel)
		go func() {
			time.Sleep(5000 * time.Millisecond)
			hasSucceededFirstLeaderElectionChannel <- false
		}()
		hasSucceededFirstLeaderElection := <-hasSucceededFirstLeaderElectionChannel
		if !hasSucceededFirstLeaderElection {
			internalLogger.LogAndReportUnexpectedError("leaderElection: first leader election failed")
			le.isLeader = true
		}
		return le
	}
}

func (le *LeaderElection) IsLeader() bool {
	return le.isLeader
}

func (le *LeaderElection) listenForChangesInLeader(hasSucceededFirstLeaderElectionChannel chan bool) {
	uniquePodName := os.Getenv(enums.PodName)
	if uniquePodName == "" {
		hasSucceededFirstLeaderElectionChannel <- false
		le.logger.LogAndReportUnexpectedError(fmt.Sprintf("env variable %s is not set", enums.PodName))
		return
	}
	namespace := os.Getenv(enums.Namespace)
	if namespace == "" {
		hasSucceededFirstLeaderElectionChannel <- false
		le.logger.LogAndReportUnexpectedError(fmt.Sprintf("env variable %s is not set", enums.Namespace))
		return
	}

	// call cancel() on terminations, to release the leader lock
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		le.logger.LogInfo("Received termination, signaling shutdown")
		cancel()
	}()

	// create the leader election config
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "datree-webhook-server-lease",
			Namespace: namespace,
		},
		Client: *le.k8sClientLeaseGetter,
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: uniquePodName,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   12 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				le.isLeader = true
				le.logger.LogInfo(fmt.Sprintf("leader election won for %s", uniquePodName))
			},
			OnStoppedLeading: func() {
				le.isLeader = false
				le.logger.LogInfo(fmt.Sprintf("leader election lost for %s", uniquePodName))
			},
			OnNewLeader: func(identity string) {
				hasSucceededFirstLeaderElectionChannel <- true
			},
		},
	})
}
