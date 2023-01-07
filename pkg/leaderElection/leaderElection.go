package leaderElection

import (
	"context"
	"errors"
	"fmt"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/datreeio/admission-webhook-datree/pkg/logger"
	v1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
	"os"
	"os/signal"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type LeaderElection struct {
	k8sClientLeaseGetter *v1.LeasesGetter
	internalLogger       logger.Logger
	isLeader             bool
}

func New(k8sClientLeaseGetter *v1.LeasesGetter, internalLogger logger.Logger) *LeaderElection {
	if k8sClientLeaseGetter == nil {
		internalLogger.LogAndReportUnexpectedError("leaderElection: k8s client is nil")
		return &LeaderElection{
			k8sClientLeaseGetter: nil,
			internalLogger:       internalLogger,
			isLeader:             true,
		}
	} else {
		le := &LeaderElection{
			k8sClientLeaseGetter: k8sClientLeaseGetter,
			internalLogger:       internalLogger,
			isLeader:             false,
		}
		// this function call is blocking, therefore we run it in a goroutine
		go le.init()
		return le
	}
}

func (le LeaderElection) IsLeader() bool {
	return le.isLeader
}

func (le LeaderElection) init() {
	uniquePodName := os.Getenv(enums.PodName)
	namespace := os.Getenv(enums.Namespace)
	if uniquePodName == "" {
		panic(errors.New(fmt.Sprintf("env variable %s is not set", enums.PodName)))
	}
	if namespace == "" {
		panic(errors.New(fmt.Sprintf("env variable %s is not set", enums.Namespace)))
	}

	// call cancel() on terminations, to release the leader lock
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		le.internalLogger.LogInfo("Received termination, signaling shutdown")
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
				le.internalLogger.LogInfo(fmt.Sprintf("leader election won for %s", uniquePodName))
			},
			OnStoppedLeading: func() {
				le.isLeader = false
				le.internalLogger.LogInfo(fmt.Sprintf("leader election lost for %s", uniquePodName))
			},
		},
	})
}
