/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package backend provides backend service.
package backend

import (
	"context"
	"io"
	"runtime"

	"git.woa.com/bk-gse/bk-nodeman/internal/manager"
	"git.woa.com/bk-gse/bk-nodeman/internal/options"
	apiv3 "git.woa.com/bk-gse/bk-nodeman/internal/router/api-v3"
	"git.woa.com/bk-gse/bk-nodeman/internal/router/basic"
	topoStorage "git.woa.com/bk-gse/bk-nodeman/internal/storage/topo"
	"git.woa.com/bk-gse/bk-nodeman/pkg/blog"
	"git.woa.com/bk-gse/bk-nodeman/pkg/config"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/client"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/discovery"
	"git.woa.com/bk-gse/bk-nodeman/pkg/rest/ssl"
	"git.woa.com/bk-gse/bk-nodeman/pkg/runtime/gopool"
	"git.woa.com/bk-gse/bk-nodeman/pkg/thirdparty/apigw"
	"git.woa.com/bk-gse/bk-nodeman/pkg/thirdparty/cmdb"
	"github.com/gin-gonic/gin"
)

const (
	// DiscoveryNameApigw defines the name of apigateway discovery.
	DiscoveryNameApigw = "apigateway"
)

// Service defines a server that provides backend services.
// It manages the configuration, lifecycle, and various capabilities (e.g., cmdb, topo storage).
// The service's capabilities are accessed through its 'cap' field, while 'router' is used to route requests.
type Service struct {
	// conf holds the configuration for the service.
	conf *config.BackendService

	// ctx is used to control the service lifecycle (cancellation and timeouts).
	ctx context.Context

	// cancelFunc is used to cancel the service and all associated operations.
	cancelFunc context.CancelFunc

	// Note: the following fields are initialized in the Start() and could not be used in other package.
	// manager workflow management.
	manager *manager.Manager

	// topoStorage bk nodeman topo storage
	topoStorage *topoStorage.DefaultStorage

	// cmdbHandler cmdb handler
	cmdbHandler cmdb.Handler

	// router is the entry point of the service, routing requests to different capabilities.
	servers []*rest.Server

	// Note: Capability is initialized in the Start() and could not be used in other package.
	// Capability is the capability of the service.
	Capability *options.Capability
}

const (
	// TopoStorageDatabase bk node manager database name.
	TopoStorageDatabase = "bknodeman_topo"

	// TopoStorageBusinessCollection bk node manager business collection name.
	TopoStorageBusinessCollection = "business"

	// TopoStorageHostCollection bk node manager host collection name.
	TopoStorageHostCollection = "host"
)

const (
	//RouterNameHttpServer defines the name of http server router.
	RouterNameHttpServer = "http-server"
)

// NewService creates a new backend service.
func NewService(conf *config.BackendService) (*Service, error) {
	svc := &Service{
		conf: conf,
	}
	svc.ctx, svc.cancelFunc = context.WithCancel(context.Background())

	var err error
	svc.cmdbHandler, err = newCMDBHandler(conf.CMDB)
	if err != nil {
		return nil, err
	}

	svc.topoStorage = topoStorage.NewStorage(&topoStorage.StorageConfig{
		MongoDB:            conf.MongoDB,
		Database:           TopoStorageDatabase,
		BusinessCollection: TopoStorageBusinessCollection,
		HostCollection:     TopoStorageHostCollection,
	})

	svc.manager = manager.NewManager(&manager.Config{
		Redis:   conf.Redis,
		MongoDB: conf.MongoDB,
	}, svc.cmdbHandler, svc.topoStorage)

	svc.Capability = &options.Capability{

		Manager:     svc.manager,
		TopoStorage: svc.topoStorage,
		CmdbHandler: svc.cmdbHandler,
	}

	httpServer := rest.NewServer(svc.ctx, RouterNameHttpServer, conf.HTTPServer.BindIP, conf.HTTPServer.Port,
		loggerWriter{},
		rest.WithPing(),
		withApiV3(svc.Capability),
		withBasic(svc.Capability),
	)
	svc.servers = append(svc.servers, httpServer)

	return svc, nil
}

// loggerWriter implements rest.LoggerWriter.
type loggerWriter struct{}

func (l loggerWriter) InfoWriter() io.Writer {
	return blog.WriterInfo{}
}

func (l loggerWriter) ErrorWriter() io.Writer {
	return blog.WriterError{}
}

// withApiV3 load api v3.
func withApiV3(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		apiv3.Load(rg, capability)
	}
}

// withBasic load basic.
func withBasic(capability *options.Capability) rest.OptionFunc {
	return func(rg *gin.RouterGroup) {
		basic.Load(rg, capability)
	}
}

// newCMDBHandler
func newCMDBHandler(conf config.CMDB) (cmdb.Handler, error) {
	apiGwHeaderSetter := newApiGwHeaderSetter(&conf.APIGateway)
	apiGwClientCapability, err := newApiGwClientCapability(&conf.APIGateway)
	if err != nil {
		return nil, err
	}

	cmdbHandler, err := cmdb.NewHandler(apiGwClientCapability, &cmdb.Config{
		TenantID:     conf.TenantID,
		HeaderSetter: apiGwHeaderSetter,
	})
	if err != nil {
		return nil, err
	}

	return cmdbHandler, nil
}

// newApiGwClientCapability creates a new api-gateway client capability.
func newApiGwClientCapability(conf *config.APIGateway) (*client.Capability, error) {
	httpClient, err := client.NewClient(&ssl.TLSConfig{
		InsecureSkipVerify: conf.TLS.InsecureSkipVerify,
		CertFile:           conf.TLS.CertFile,
		KeyFile:            conf.TLS.KeyFile,
		CAFile:             conf.TLS.CAFile,
		Password:           conf.TLS.Password,
	})
	if err != nil {
		return nil, err
	}

	clientCap := &client.Capability{
		Client:               httpClient,
		Discover:             discovery.NewDiscovery(DiscoveryNameApigw, conf.Endpoints),
		ToleranceLatencyTime: client.ToleranceLatencyTimeDefault,
		MetricOpts:           client.MetricOption{},
		Logger:               blog.GlobalLogger{},
	}

	return clientCap, nil
}

// newApiGwHeaderSetter creates a new api-gateway header setter.
func newApiGwHeaderSetter(conf *config.APIGateway) apigw.HeaderSetter {
	return &apigw.Config{
		Endpoints:   conf.Endpoints,
		AppCode:     conf.AppCode,
		AppSecret:   conf.AppSecret,
		User:        conf.User,
		AuthMode:    apigw.AuthMode(conf.AuthMode),
		BkTicket:    conf.BkTicket,
		BkToken:     conf.BkToken,
		AccessToken: conf.AccessToken,
	}
}

// Start starts the backend service.
func (svc *Service) Start(ctx context.Context) error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logConfig := blog.NewLogConfig()
	logConfig.LogDir = svc.conf.Log.Dir
	logConfig.LogMaxSizeMB = svc.conf.Log.MaxSizeMB
	logConfig.LogMaxNum = svc.conf.Log.MaxNum
	logConfig.Level = svc.conf.Log.Level
	logConfig.ToStdErr = svc.conf.Log.ToStdErr
	logConfig.AlsoToStdErr = svc.conf.Log.AlsoToStdErr
	blog.InitLogs(logConfig)

	svc.ctx, svc.cancelFunc = context.WithCancel(ctx)

	if err := svc.topoStorage.Start(ctx); err != nil {
		blog.Errorf("failed to start topo storage, err: %v", err)

		return err
	}

	if err := svc.manager.Start(ctx); err != nil {
		blog.Errorf("failed to start manager, err: %v", err)

		return err
	}

	// start servers
	gp := gopool.NewPool()
	for _, router := range svc.servers {
		// router start will block until router stop, so we need to run it in a goroutine.
		fn := func() error {
			if err := router.Start(); err != nil {
				return err
			}

			return nil
		}
		gp.Go(fn)
		blog.Infof("start router,name: %s, ip: %s, port: %d",
			router.Name(), router.IP(), router.Port())
	}

	// wait until all servers stopped or application error.
	if err := gp.Wait(); err != nil {
		blog.Errorf("failed to start servers, err: %v", err)
		return err
	}

	return nil
}
