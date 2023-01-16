package webhookhandler

import (
	"context"
	"crypto/tls"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	v1 "nano-pod-operator/api/v1"
	"net"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sync"
	"testing"
	"time"
)

var logger = zap.New(zap.UseDevMode(true))

var (
	k8sClient  client.Client
	testEnv    *envtest.Environment
	testScheme *runtime.Scheme = scheme.Scheme
	ctx        context.Context
	cancel     context.CancelFunc
)

func TestMain(m *testing.M) {
	ctx, cancel = context.WithCancel(context.TODO())
	defer cancel()

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "webhook")},
		},
	}
	cfg, err := testEnv.Start()
	if err != nil {
		logger.Error(err, "failed to start test env.")
		os.Exit(1)
	}

	err = v1.AddToScheme(testScheme)
	if err != nil {
		logger.Error(err, "failed to register scheme.")
		os.Exit(1)
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: testScheme})
	if err != nil {
		logger.Error(err, "failed to setup client.")
		os.Exit(1)
	}

	// start webhook server using Manager
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             testScheme,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	if err != nil {
		logger.Error(err, "failed to Start webhook server.")
		os.Exit(1)
	}

	err = (&v1.NanoPod{}).SetupWebhookWithManager(mgr)
	if err != nil {
		logger.Error(err, "failed to SetupWebhookWithManager.")
		os.Exit(1)
	}

	ctx, cancel = context.WithCancel(context.TODO())
	defer cancel()
	go func() {
		err = mgr.Start(ctx)
		if err != nil {
			logger.Error(err, "failed to start manager.")
			os.Exit(1)
		}
	}()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err = retry.OnError(wait.Backoff{
			Steps:    20,
			Duration: 10 * time.Millisecond,
			Factor:   1.5,
			Jitter:   0.1,
			Cap:      time.Second * 30,
		}, func(error) bool {
			return true
		}, func() error {
			conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
			if err != nil {
				return err
			}
			_ = conn.Close()
			return nil
		})
		if err != nil {
			logger.Error(err, "failed to wait for webhook server to be ready.")
			os.Exit(1)
		}
	}(wg)

	wg.Wait()

	code := m.Run()

	err = testEnv.Stop()
	if err != nil {
		logger.Error(err, "failed to stop test env.")
		os.Exit(1)
	}

	os.Exit(code)

}
