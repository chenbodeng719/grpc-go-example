## etcd

- cmd
```
etcd --name infra2 --listen-client-urls http://127.0.0.1:22379 --advertise-client-urls http://127.0.0.1:22379 --listen-peer-urls http://127.0.0.1:22380 --initial-advertise-peer-urls http://127.0.0.1:22380  --enable-pprof --logger=zap --log-outputs=stderr

etcdctl --endpoints=http://127.0.0.1:22379 get hello --prefix

```