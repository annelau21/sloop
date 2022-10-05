/*
 * Copyright (c) 2019, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see LICENSE.txt file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */

package server

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/salesforce/sloop/pkg/sloop/ingress"
	"github.com/salesforce/sloop/pkg/sloop/server/internal/config"
	"github.com/salesforce/sloop/pkg/sloop/store/typed"
	"os"
	"strings"
)

const alsologtostderr = "alsologtostderr"

func RealMain() error {
	defer glog.Flush()
	setupStdErrLogging()

	conf := config.Init() // internally this calls flag.parse
	glog.Infof("SloopConfig: %v", conf.ToYaml())

	err := conf.Validate()
	if err != nil {
		return errors.Wrap(err, "config validation failed")
	}

	kubeContext, err := ingress.GetKubernetesContext(conf.ApiServerHost, conf.UseKubeContext, conf.PrivilegedAccess)
	if err != nil {
		return errors.Wrap(err, "failed to get kubernetes context")
	}

	// Channel used for updates from ingress to store
	// The channel is owned by this function, and no external code should close this!
	kubeWatchChan := make(chan typed.KubeWatchResult, 1000)

	//factory := &badgerwrap.BadgerFactory{}
	//
	//storeRootWithKubeContext := path.Join(conf.StoreRoot, kubeContext)
	//storeConfig := &untyped.Config{
	//	RootPath:                 storeRootWithKubeContext,
	//	ConfigPartitionDuration:  time.Duration(1) * time.Hour,
	//	BadgerMaxTableSize:       conf.BadgerMaxTableSize,
	//	BadgerKeepL0InMemory:     conf.BadgerKeepL0InMemory,
	//	BadgerVLogFileSize:       conf.BadgerVLogFileSize,
	//	BadgerVLogMaxEntries:     conf.BadgerVLogMaxEntries,
	//	BadgerUseLSMOnlyOptions:  conf.BadgerUseLSMOnlyOptions,
	//	BadgerEnableEventLogging: conf.BadgerEnableEventLogging,
	//	BadgerNumOfCompactors:    conf.BadgerNumOfCompactors,
	//	BadgerNumL0Tables:        conf.BadgerNumL0Tables,
	//	BadgerNumL0TablesStall:   conf.BadgerNumL0TablesStall,
	//	BadgerSyncWrites:         conf.BadgerSyncWrites,
	//	BadgerLevelOneSize:       conf.BadgerLevelOneSize,
	//	BadgerLevSizeMultiplier:  conf.BadgerLevSizeMultiplier,
	//	BadgerVLogFileIOMapping:  conf.BadgerVLogFileIOMapping,
	//	BadgerVLogTruncate:       conf.BadgerVLogTruncate,
	//	BadgerDetailLogEnabled:   conf.BadgerDetailLogEnabled,
	//}
	//db, err := untyped.OpenStore(factory, storeConfig)
	//if err != nil {
	//	return errors.Wrap(err, "failed to init untyped store")
	//}
	//defer untyped.CloseStore(db)
	//
	//if conf.RestoreDatabaseFile != "" {
	//	glog.Infof("Restoring from backup file %q into context %q", conf.RestoreDatabaseFile, kubeContext)
	//	err := ingress.DatabaseRestore(db, conf.RestoreDatabaseFile)
	//	if err != nil {
	//		return errors.Wrap(err, "failed to restore database")
	//	}
	//	glog.Infof("Restored from backup file %q into context %q", conf.RestoreDatabaseFile, kubeContext)
	//}
	//
	//tables := typed.NewTableList(db)
	//processor := processing.NewProcessing(kubeWatchChan, tables, conf.KeepMinorNodeUpdates, conf.MaxLookback)
	//processor.Start()

	manyInserts := false
	listOfKubeContextMap := map[string]string{"ast-sam": "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/ast-sam/",
		"aws-dev2-uswest2-apiq-sam-processing1":               "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-apiq-sam-processing1/",
		"aws-dev2-uswest2-apiq-sam-restricted1":               "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-apiq-sam-restricted1/",
		"aws-dev2-uswest2-ast-sam-processing1":                "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sam-processing1/",
		"aws-dev2-uswest2-ast-sfci-gec1":                      "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sfci-gec1/",
		"aws-dev2-uswest2-ast-sfci-gecteam1":                  "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sfci-gecteam1/",
		"aws-dev2-uswest2-ast-sfci-release1":                  "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sfci-release1/",
		"aws-dev2-uswest2-ast-sfci-team1":                     "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sfci-team1/",
		"aws-dev2-uswest2-buildndeliver-armada2":              "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-buildndeliver-armada2/",
		"aws-dev2-uswest2-cdp1-cdp-dev2-1":                    "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-cdp1-cdp-dev2-1/",
		"aws-dev2-uswest2-controltelemetry-fsre-shared-eks1":  "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-controltelemetry-fsre-shared-eks1/",
		"aws-dev2-uswest2-controltelemetry-sam-processing1":   "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-controltelemetry-sam-processing1/",
		"aws-dev2-uswest2-dp-services-sam-processing1":        "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-dp-services-sam-processing1/",
		"aws-dev2-uswest2-foundation-fsre-shared-eks2":        "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-foundation-fsre-shared-eks2/",
		"aws-dev2-uswest2-foundation-sam-logging-monitoring1": "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-foundation-sam-logging-monitoring1/",
		"aws-dev2-uswest2-foundation-sam-processing1":         "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-foundation-sam-processing1/",
		"aws-dev2-uswest2-messagingjourneys1-sam-processing1": "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-messagingjourneys1-sam-processing1/",
		"aws-dev2-uswest2-messagingjourneys1-sam-processing2": "https://pseudo-kubeapi.sfproxy.core3.test1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-messagingjourneys1-sam-processing2/"}
	// Real kubernetes watcher
	var kubeWatcherSource ingress.KubeWatcher
	if !conf.DisableKubeWatcher {
		if !manyInserts {
			kubeClient, err := ingress.MakeKubernetesClient(conf.ApiServerHost, kubeContext, conf.PrivilegedAccess)
			if err != nil {
				return errors.Wrap(err, "failed to create kubernetes client")
			}

			kubeWatcherSource, err = ingress.NewKubeWatcherSource(kubeClient, kubeWatchChan, conf.KubeWatchResyncInterval, conf.WatchCrds, conf.CrdRefreshInterval, conf.ApiServerHost, kubeContext, conf.EnableGranularMetrics)
			if err != nil {
				return errors.Wrap(err, "failed to initialize kubeWatcher")
			}
		}

		if manyInserts {
			for kC, masterURL := range listOfKubeContextMap {
				fmt.Sprintf("kubeContext is %v", kC)
				kubeClient, err := ingress.MakeKubernetesClient(masterURL, kC, false)
				if err != nil {
					return errors.Wrap(err, "failed to create kubernetes client")
				}

				kubeWatcherSource, err = ingress.NewKubeWatcherSource(kubeClient, kubeWatchChan, conf.KubeWatchResyncInterval, conf.WatchCrds, conf.CrdRefreshInterval, masterURL, kC, conf.EnableGranularMetrics)
				if err != nil {
					return errors.Wrap(err, "failed to initialize kubeWatcher")
				}
			}
		}

	}

	////File playback
	//if conf.DebugPlaybackFile != "" {
	//	err = ingress.PlayFile(kubeWatchChan, conf.DebugPlaybackFile)
	//	if err != nil {
	//		return errors.Wrap(err, "failed to play back file")
	//	}
	//}
	//
	//var recorder *ingress.FileRecorder
	//if conf.DebugRecordFile != "" {
	//	recorder = ingress.NewFileRecorder(conf.DebugRecordFile, kubeWatchChan)
	//	recorder.Start()
	//}
	//
	//var storemgr *storemanager.StoreManager
	//if !conf.DisableStoreManager {
	//	fs := &afero.Afero{Fs: afero.NewOsFs()}
	//	storeCfg := &storemanager.Config{
	//		StoreRoot:          conf.StoreRoot,
	//		Freq:               conf.CleanupFrequency,
	//		TimeLimit:          conf.MaxLookback,
	//		SizeLimitBytes:     conf.MaxDiskMb * 1024 * 1024,
	//		BadgerDiscardRatio: conf.BadgerDiscardRatio,
	//		BadgerVLogGCFreq:   conf.BadgerVLogGCFreq,
	//		DeletionBatchSize:  conf.DeletionBatchSize,
	//		GCThreshold:        conf.ThresholdForGC,
	//		EnableDeleteKeys:   conf.EnableDeleteKeys,
	//	}
	//	storemgr = storemanager.NewStoreManager(tables, storeCfg, fs)
	//	storemgr.Start()
	//}
	//
	//displayContext := kubeContext
	//if conf.DisplayContext != "" {
	//	displayContext = conf.DisplayContext
	//}
	//
	//webConfig := webserver.WebConfig{
	//	BindAddress:      conf.BindAddress,
	//	Port:             conf.Port,
	//	WebFilesPath:     conf.WebFilesPath,
	//	ConfigYaml:       conf.ToYaml(),
	//	MaxLookback:      conf.MaxLookback,
	//	DefaultNamespace: conf.DefaultNamespace,
	//	DefaultLookback:  conf.DefaultLookback,
	//	DefaultResources: conf.DefaultKind,
	//	ResourceLinks:    conf.ResourceLinks,
	//	LeftBarLinks:     conf.LeftBarLinks,
	//	CurrentContext:   displayContext,
	//}
	//err = webserver.Run(webConfig, tables)
	//if err != nil {
	//	return errors.Wrap(err, "failed to run webserver")
	//}

	// Initiate shutdown with the following order:
	// 1. Shut down ingress so that it stops emitting events
	// 2. Close the input channel which signals processing to finish work
	// 3. Wait on processor to tell us all work is complete.  Store will not change after that
	if kubeWatcherSource != nil {
		kubeWatcherSource.Stop()
	}
	close(kubeWatchChan)
	//processor.Wait()
	//
	//if recorder != nil {
	//	recorder.Close()
	//}
	//
	//if storemgr != nil {
	//	storemgr.Shutdown()
	//}

	glog.Infof("RunWithConfig finished")
	return nil
}

// By default glog will not print anything to console, which can confuse users
// This will turn it on unless user sets it explicitly (with --alsologtostderr=false)
func setupStdErrLogging() {
	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, alsologtostderr) {
			return
		}
	}
	err := flag.Set("alsologtostderr", "true")
	if err != nil {
		panic(err)
	}
}
