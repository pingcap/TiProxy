// Copyright 2024 PingCAP, Inc.
// SPDX-License-Identifier: Apache-2.0

package elect

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pingcap/tiproxy/lib/util/errors"
	"github.com/pingcap/tiproxy/lib/util/retry"
	"github.com/pingcap/tiproxy/lib/util/waitgroup"
	"github.com/pingcap/tiproxy/pkg/metrics"
	"github.com/pingcap/tiproxy/pkg/util/etcd"
	"github.com/siddontang/go/hack"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
)

const (
	logInterval    = 10
	ownerKeyPrefix = "/tiproxy/"
	ownerKeySuffix = "/owner"
)

type Member interface {
	OnElected()
	OnRetired()
}

// Election is used to campaign the owner and manage the owner information.
type Election interface {
	// Start starts compaining the owner.
	Start(context.Context)
	// ID returns the member ID.
	ID() string
	// IsOwner returns whether the member is the owner.
	IsOwner() bool
	// GetOwnerID gets the owner ID.
	GetOwnerID(ctx context.Context) (string, error)
	// Close stops compaining the owner.
	Close()
}

type ElectionConfig struct {
	Timeout    time.Duration
	RetryIntvl time.Duration
	RetryCnt   uint64
	SessionTTL int
}

func DefaultElectionConfig(sessionTTL int) ElectionConfig {
	return ElectionConfig{
		Timeout:    2 * time.Second,
		RetryIntvl: 500 * time.Millisecond,
		RetryCnt:   3,
		SessionTTL: sessionTTL,
	}
}

var _ Election = (*election)(nil)

// election is used for electing owner.
type election struct {
	cfg ElectionConfig
	// id is typically the instance address
	id  string
	key string
	// trimedKey is shown as a label in grafana
	trimedKey string
	lg        *zap.Logger
	etcdCli   *clientv3.Client
	elec      atomic.Pointer[concurrency.Election]
	wg        waitgroup.WaitGroup
	cancel    context.CancelFunc
	member    Member
}

// NewElection creates an Election.
func NewElection(lg *zap.Logger, etcdCli *clientv3.Client, cfg ElectionConfig, id, key string, member Member) *election {
	lg = lg.With(zap.String("key", key), zap.String("id", id))
	return &election{
		lg:        lg,
		etcdCli:   etcdCli,
		cfg:       cfg,
		id:        id,
		key:       key,
		trimedKey: strings.TrimSuffix(strings.TrimPrefix(key, ownerKeyPrefix), ownerKeySuffix),
		member:    member,
	}
}

func (m *election) Start(ctx context.Context) {
	// No PD.
	if m.etcdCli == nil {
		return
	}
	clientCtx, cancelFunc := context.WithCancel(ctx)
	m.cancel = cancelFunc
	// Don't recover because we don't know what will happen after recovery.
	m.wg.Run(func() {
		m.campaignLoop(clientCtx)
	})
}

func (m *election) ID() string {
	return m.id
}

func (m *election) initSession(ctx context.Context) (*concurrency.Session, error) {
	var session *concurrency.Session
	// If the network breaks for sometime, the session will fail but it still needs to compaign after recovery.
	// So retry it infinitely.
	err := retry.RetryNotify(func() error {
		var err error
		// Do not use context.WithTimeout, otherwise the session will be cancelled after timeout, even if it is created successfully.
		session, err = concurrency.NewSession(m.etcdCli, concurrency.WithTTL(m.cfg.SessionTTL), concurrency.WithContext(ctx))
		return err
	}, ctx, m.cfg.RetryIntvl, retry.InfiniteCnt,
		func(err error, duration time.Duration) {
			m.lg.Warn("failed to init election session, retrying", zap.Error(err))
		}, logInterval)
	if err == nil {
		m.lg.Info("election session is initialized")
	} else {
		m.lg.Error("failed to init election session, quit", zap.Error(err))
	}
	return session, err
}

func (m *election) IsOwner() bool {
	ownerID, err := m.GetOwnerID(context.Background())
	if err != nil {
		return false
	}
	return ownerID == m.id
}

func (m *election) campaignLoop(ctx context.Context) {
	session, err := m.initSession(ctx)
	if err != nil {
		return
	}
	for {
		select {
		case <-session.Done():
			m.lg.Info("etcd session is done, creates a new one")
			leaseID := session.Lease()
			if session, err = m.initSession(ctx); err != nil {
				m.lg.Error("new session failed, break campaign loop", zap.Error(err))
				m.revokeLease(leaseID)
				return
			}
		case <-ctx.Done():
			m.revokeLease(session.Lease())
			return
		default:
		}
		// If the etcd server turns clocks forward, the following case may occur.
		// The etcd server deletes this session's lease ID, but etcd session doesn't find it.
		// In this case if we do the campaign operation, the etcd server will return ErrLeaseNotFound.
		if errors.Is(err, rpctypes.ErrLeaseNotFound) {
			if session != nil {
				err = session.Close()
				m.lg.Warn("etcd session encounters ErrLeaseNotFound, close it", zap.Error(err))
			}
			continue
		}

		elec := concurrency.NewElection(session, m.key)
		if err = elec.Campaign(ctx, m.id); err != nil {
			m.lg.Info("failed to campaign", zap.Error(err))
			continue
		}

		ownerID, err := m.GetOwnerID(ctx)
		if err != nil || ownerID != m.id {
			continue
		}

		m.onElected(elec)
		// NOTICE: watchOwner won't revoke the lease.
		m.watchOwner(ctx, session, ownerID)
		m.onRetired()
	}
}

func (m *election) onElected(elec *concurrency.Election) {
	m.member.OnElected()
	m.elec.Store(elec)
	metrics.OwnerGauge.WithLabelValues(m.trimedKey).Set(1)
	m.lg.Info("elected as the owner")
}

func (m *election) onRetired() {
	m.member.OnRetired()
	m.elec.Store(nil)
	// Delete the metric so that it doesn't show on Grafana.
	metrics.OwnerGauge.MetricVec.DeletePartialMatch(map[string]string{metrics.LblType: m.trimedKey})
	m.lg.Info("the owner retires")
}

// revokeLease revokes the session lease so that other members can compaign immediately.
func (m *election) revokeLease(leaseID clientv3.LeaseID) {
	// If revoke takes longer than the ttl, lease is expired anyway.
	// Don't use the context of the caller because it may be already done.
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Duration(m.cfg.SessionTTL)*time.Second)
	if _, err := m.etcdCli.Revoke(cancelCtx, leaseID); err != nil {
		m.lg.Warn("revoke session failed", zap.Error(errors.WithStack(err)))
	}
	cancel()
}

// GetOwnerID is similar to concurrency.Election.Leader() but it doesn't need an concurrency.Election.
func (m *election) GetOwnerID(ctx context.Context) (string, error) {
	if m.etcdCli == nil {
		return "", concurrency.ErrElectionNoLeader
	}
	kvs, err := etcd.GetKVs(ctx, m.etcdCli, m.key, clientv3.WithFirstCreate(), m.cfg.Timeout, m.cfg.RetryIntvl, m.cfg.RetryCnt)
	if err != nil {
		return "", err
	}
	if len(kvs) == 0 {
		return "", concurrency.ErrElectionNoLeader
	}
	return hack.String(kvs[0].Value), nil
}

func (m *election) watchOwner(ctx context.Context, session *concurrency.Session, key string) {
	watchCh := m.etcdCli.Watch(ctx, key)
	for {
		select {
		case resp, ok := <-watchCh:
			if !ok {
				m.lg.Info("watcher is closed, no owner")
				return
			}
			if resp.Canceled {
				m.lg.Info("watch canceled, no owner")
				return
			}

			for _, ev := range resp.Events {
				if ev.Type == mvccpb.DELETE {
					m.lg.Info("watch failed, owner is deleted")
					return
				}
			}
		case <-session.Done():
			return
		case <-ctx.Done():
			return
		}
	}
}

// Close is called before the instance is going to shutdown.
// It should hand over the owner to someone else.
func (m *election) Close() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.wg.Wait()
}
