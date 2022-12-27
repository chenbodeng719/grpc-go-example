package endpoint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.uber.org/zap"
	"log"
	"strings"
	"time"
)

type Register struct {
	EtcdAddrs   []string
	DialTimeout int

	closeCh     chan struct{}
	leasesID    clientv3.LeaseID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse

	manager endpoints.Manager
	srvInfo Server
	srvTTL  int64
	cli     *clientv3.Client
	logger  *zap.Logger
}

func NewRegister(etcdAddrs []string, logger *zap.Logger) *Register {
	return &Register{
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
	}
}

// Register a service
func (r *Register) Register(srvInfo Server, ttl int64) (chan<- struct{}, error) {
	var err error

	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return nil, errors.New("invalid ip")
	}

	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return nil, err
	}

	r.srvInfo = srvInfo
	r.srvTTL = ttl

	r.manager, err = endpoints.NewManager(r.cli, srvInfo.Name)
	if err != nil {
		return nil, err
	}
	if err = r.register(); err != nil {
		return nil, err
	}

	r.closeCh = make(chan struct{})

	go r.keepAlive()

	return r.closeCh, nil
}

// register 注册节点
func (r *Register) register() error {
	leaseCtx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	leaseResp, err := r.cli.Grant(leaseCtx, r.srvTTL)
	if err != nil {
		return err
	}
	r.leasesID = leaseResp.ID
	if r.keepAliveCh, err = r.cli.KeepAlive(leaseCtx, leaseResp.ID); err != nil {
		return err
	}

	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}
	log.Println("fuck", BuildRegPath(r.srvInfo), string(data))

	cerr := r.manager.AddEndpoint(leaseCtx, BuildRegPath(r.srvInfo), endpoints.Endpoint{Addr: r.srvInfo.Addr}, clientv3.WithLease(r.leasesID))
	//_, err = r.cli.Put(leaseCtx, BuildRegPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
	return cerr
}

// keepAlive
func (r *Register) keepAlive() {
	//ticker := time.NewTicker(time.Duration(r.srvTTL) * time.Second)
	for {
		select {
		case <-r.closeCh:
			fmt.Println("close")
			if err := r.unregister(); err != nil {
				r.logger.Error("unregister failed", zap.Error(err))
			}
			if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
				r.logger.Error("revoke failed", zap.Error(err))
			}
			//case res := <-r.keepAliveCh:
			//	if res == nil {
			//		if err := r.register(); err != nil {
			//			r.logger.Error("register failed", zap.Error(err))
			//		}
			//	}
			//case <-ticker.C:
			//	//log.Println("ticker")
			//	if r.keepAliveCh == nil {
			//		if err := r.register(); err != nil {
			//			r.logger.Error("register failed", zap.Error(err))
			//		}
			//	}
		}
	}
	return
}

// unregister 删除节点
func (r *Register) unregister() error {
	//_, err := r.cli.Delete(context.Background(), BuildRegPath(r.srvInfo))
	err := r.manager.DeleteEndpoint(context.Background(), BuildRegPath(r.srvInfo))
	return err
}

func (r *Register) Stop() {
	r.closeCh <- struct{}{}
}

func BuildRegPath(info Server) string {
	return fmt.Sprintf("%s/%s", info.Name, info.Addr)
	//return fmt.Sprintf("%s", info.Name)
}
