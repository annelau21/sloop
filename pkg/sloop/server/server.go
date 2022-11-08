/*
 * Copyright (c) 2019, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see LICENSE.txt file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */

package server

import (
	"flag"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/salesforce/sloop/pkg/sloop/ingress"
	"github.com/salesforce/sloop/pkg/sloop/processing"
	"github.com/salesforce/sloop/pkg/sloop/server/internal/config"
	"github.com/salesforce/sloop/pkg/sloop/store/typed"
	"github.com/salesforce/sloop/pkg/sloop/store/untyped"
	"github.com/salesforce/sloop/pkg/sloop/store/untyped/badgerwrap"
	"github.com/salesforce/sloop/pkg/sloop/storemanager"
	"github.com/salesforce/sloop/pkg/sloop/webserver"
	"github.com/spf13/afero"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var listOfKubeContextMap1 = []string{"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/ast-sam/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-apiq-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-apiq-sam-restricted1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sfci-gec1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sfci-gecteam1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sfci-release1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-ast-sfci-team1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-buildndeliver-armada2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-cdp1-cdp-dev2-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-controltelemetry-fsre-shared-eks1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-controltelemetry-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-dp-services-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-foundation-fsre-shared-eks2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-foundation-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-foundation-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-messagingjourneys1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-messagingjourneys1-sam-processing2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-seas-dev2-eks-usw2-seas-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-seas-dev2-eks-usw2-seas-2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-seas-dev2-eks-usw2-seas-3/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-seas-dev2-eks-usw2-seas-4/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-seas-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev2-uswest2-tvm-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-autobuilds-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-einstein-dev4-eks-usw2-compute-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-einstein-dev4-eks-usw2-compute-2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-einstein-dev4-eks-usw2-infra-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-einstein-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-fieldservice-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-foundation-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-foundation-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-industries-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-mobilebuild1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-mulesoft-cp-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-mulesoft-regional1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-nft1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-armada2dev/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-hdpsampekscluster001/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-sam-processing-hccptest1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-sam-processing1-2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-sam-processing1-3/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-sam-processing5/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-sam-processing6/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad4-sam-processing7/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-scratchpad6-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-dev4-uswest2-tableau-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/aws-giadev1-usgoveast1-core1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-analyticsdataservice-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-buildndeliver-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp001-cdp-data-capture-dev1-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp001-cdp-dev1-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp001-cdp-dev1-1-dpc1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp001-datorama1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp001-mce-dev1-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp001-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp001-sam-restricted1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp002-cdp-data-capture-dev1-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp002-cdp-dev1-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp002-cdp-dev1-1-dpc1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp002-datorama1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-cdp002-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-commerceecom-sam-processing0001/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-pc-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-pc-hbase2a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-pc-hbase3a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-pc-hbase4a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-phoenix-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-phoenix-hbase2a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-phoenix-hbase4a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-phoenix-hbase5a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-phoenix-hbase5b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-sh-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-sor-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-sor-hbase1b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-sor-hbase2a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-dev-sor-hbase2b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-hbase1b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-hbase2a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-hbase2b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-pc3/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-sam-processing00001/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-sam-processing00002/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-sam-restricted1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-sam-restricted2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-sdb-services-test15/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-core002-sdb-services-test6/"}

var listOfKubeContextMap2 = []string{
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-coretip1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-dbaas-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-devpcs-devpcs-services/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-devpcs-devpcs-ui/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-diffeo-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-dnr-dnr-airflow-k8s1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-dnr-dnr-spark-k8s1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-edge-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-einstein-dev-eks-usw2-hawking-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-einstein-dev-eks-usw2-hawking-2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-einstein-dev-eks-usw2-hawking-5/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-einstein-dev1-eks-usw2-compute-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-einstein-dev1-eks-usw2-compute-3/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-einstein-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-einstein-sam-restricted1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-foundation-sam-control1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-foundation-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-foundation-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-gid-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-ci-stg-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-ci-stg-hbase1b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-dev-hgrate-hbase2a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-dev-hgrate-hbase2b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-dev-perf-hbase4a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-dev-perf-hbase4b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-dev-phoenix-hbase3a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-hbase1b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-hbase3a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-monitoring-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-netsec-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-pcs-pcs-services1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-pcs-pcs-ui1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-scratchpad2-sfci-release1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-scratchpad2-sfci-team1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-scratchpad3-fre-scratchpad3/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-scratchpad3-pcs1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-scratchpad3-pcs2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-seas-dev-eks-usw2-seas-10/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-seas-dev-eks-usw2-seas-6/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-seas-dev-eks-usw2-seas-7/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-seas-dev-eks-usw2-seas-8/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-seas-dev-eks-usw2-seas-9/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-secrets-htha-test-vault/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-tvm-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-uip001-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/dev1-uswest2-unified-engagement-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/fdev1-uswest2-dnr-dnr-airflow-k8s1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/fdev1-uswest2-dnr-dnr-spark-k8s1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/fdev1-uswest2-foundation-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/fdev1-uswest2-foundation-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/fdev1-uswest2-monitoring-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/fdev1-uswest2-monitoring-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/fdev1-uswest2-monitoring-sam-logging-monitoring2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/fdev1-uswest2-monitoring-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-core1-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-core1-hbase1b/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-core1-hbase2a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-core1-hbase3a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-core1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-core1-sam-processing2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-core1-sdb-services-prod/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-diffeo-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-dp-services-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-edge-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-foundation-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-foundation-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-gid-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-industries-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-monitoring-hbase1a/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-monitoring-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-monitoring-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-unified-engagement1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf1-useast2-unified-engagement1-sam-restricted1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-cdp1-cdp-data-capture-perf2-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-cdp1-cdp-perf2-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-cdp1-cdp-perf2-1-dpc1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-cdp1-mce-perf2-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-cdp1-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-cdp1-sam-restricted1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-dp-services-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-einstein-perf2-eks-usw2-compute-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-einstein-perf2-eks-usw2-compute-2/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-einstein-perf2-eks-usw2-infra-1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-einstein-sam-processing1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-foundation-sam-logging-monitoring1/",
	"https://pseudo-kubeapi.sfproxy.core002.dev1-uswest2.aws.sfdc.cl/perf2-uswest2-foundation-sam-processing1/"}

const alsologtostderr = "alsologtostderr"

func RealMain2(group int) error {
	defer glog.Flush()
	setupStdErrLogging()
	var listOfKubeContextMap []string
	conf := config.Init() // internally this calls flag.parse
	glog.Infof("SloopConfig: %v", conf.ToYaml())

	kubeWatchChan := make(chan typed.KubeWatchResult, 1000)

	var kubeWatcherSource ingress.KubeWatcher
	var wg sync.WaitGroup
	// Determine which group to be
	switch group {
	case 1:
		listOfKubeContextMap = listOfKubeContextMap1
	case 2:
		listOfKubeContextMap = listOfKubeContextMap2
	default:
		listOfKubeContextMap = listOfKubeContextMap1
	}

	for _, masterURL := range listOfKubeContextMap {
		wg.Add(1)
		go func(masterURL string) {
			defer wg.Done()
			//fmt.Sprintf("kubeContext is %v", kC)
			kubeClient, err := ingress.MakeKubernetesClient(masterURL, "", true)
			if err != nil {
				glog.Infof("failed to create kubernetes client")
			}

			kubeWatcherSource, err = ingress.NewKubeWatcherSource(kubeClient, kubeWatchChan, conf.KubeWatchResyncInterval, conf.WatchCrds, conf.CrdRefreshInterval, masterURL, "", conf.EnableGranularMetrics)
			if err != nil {
				glog.Infof("failed to initialize kubeWatcher")
			}
		}(masterURL)
		if kubeWatcherSource != nil {
			kubeWatcherSource.Stop()
		}
		time.Sleep(10)
	}
	wg.Wait()

	glog.Infof("RunWithConfig finished")
	return nil
}

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

	factory := &badgerwrap.BadgerFactory{}

	storeRootWithKubeContext := path.Join(conf.StoreRoot, kubeContext)
	storeConfig := &untyped.Config{
		RootPath:                 storeRootWithKubeContext,
		ConfigPartitionDuration:  time.Duration(1) * time.Hour,
		BadgerMaxTableSize:       conf.BadgerMaxTableSize,
		BadgerKeepL0InMemory:     conf.BadgerKeepL0InMemory,
		BadgerVLogFileSize:       conf.BadgerVLogFileSize,
		BadgerVLogMaxEntries:     conf.BadgerVLogMaxEntries,
		BadgerUseLSMOnlyOptions:  conf.BadgerUseLSMOnlyOptions,
		BadgerEnableEventLogging: conf.BadgerEnableEventLogging,
		BadgerNumOfCompactors:    conf.BadgerNumOfCompactors,
		BadgerNumL0Tables:        conf.BadgerNumL0Tables,
		BadgerNumL0TablesStall:   conf.BadgerNumL0TablesStall,
		BadgerSyncWrites:         conf.BadgerSyncWrites,
		BadgerLevelOneSize:       conf.BadgerLevelOneSize,
		BadgerLevSizeMultiplier:  conf.BadgerLevSizeMultiplier,
		BadgerVLogFileIOMapping:  conf.BadgerVLogFileIOMapping,
		BadgerVLogTruncate:       conf.BadgerVLogTruncate,
		BadgerDetailLogEnabled:   conf.BadgerDetailLogEnabled,
	}
	db, err := untyped.OpenStore(factory, storeConfig)
	if err != nil {
		return errors.Wrap(err, "failed to init untyped store")
	}
	defer untyped.CloseStore(db)

	if conf.RestoreDatabaseFile != "" {
		glog.Infof("Restoring from backup file %q into context %q", conf.RestoreDatabaseFile, kubeContext)
		err := ingress.DatabaseRestore(db, conf.RestoreDatabaseFile)
		if err != nil {
			return errors.Wrap(err, "failed to restore database")
		}
		glog.Infof("Restored from backup file %q into context %q", conf.RestoreDatabaseFile, kubeContext)
	}

	tables := typed.NewTableList(db)
	processor := processing.NewProcessing(kubeWatchChan, tables, conf.KeepMinorNodeUpdates, conf.MaxLookback)
	processor.Start()

	// Real kubernetes watcher
	var kubeWatcherSource ingress.KubeWatcher
	if !conf.DisableKubeWatcher {
		kubeClient, err := ingress.MakeKubernetesClient(conf.ApiServerHost, "", conf.PrivilegedAccess)
		if err != nil {
			return errors.Wrap(err, "failed to create kubernetes client")
		}

		kubeWatcherSource, err = ingress.NewKubeWatcherSource(kubeClient, kubeWatchChan, conf.KubeWatchResyncInterval, conf.WatchCrds, conf.CrdRefreshInterval, conf.ApiServerHost, "", conf.EnableGranularMetrics)
		if err != nil {
			return errors.Wrap(err, "failed to initialize kubeWatcher")
		}

	}

	//File playback
	if conf.DebugPlaybackFile != "" {
		err = ingress.PlayFile(kubeWatchChan, conf.DebugPlaybackFile)
		if err != nil {
			return errors.Wrap(err, "failed to play back file")
		}
	}

	var recorder *ingress.FileRecorder
	if conf.DebugRecordFile != "" {
		recorder = ingress.NewFileRecorder(conf.DebugRecordFile, kubeWatchChan)
		recorder.Start()
	}

	var storemgr *storemanager.StoreManager
	if !conf.DisableStoreManager {
		fs := &afero.Afero{Fs: afero.NewOsFs()}
		storeCfg := &storemanager.Config{
			StoreRoot:          conf.StoreRoot,
			Freq:               conf.CleanupFrequency,
			TimeLimit:          conf.MaxLookback,
			SizeLimitBytes:     conf.MaxDiskMb * 1024 * 1024,
			BadgerDiscardRatio: conf.BadgerDiscardRatio,
			BadgerVLogGCFreq:   conf.BadgerVLogGCFreq,
			DeletionBatchSize:  conf.DeletionBatchSize,
			GCThreshold:        conf.ThresholdForGC,
			EnableDeleteKeys:   conf.EnableDeleteKeys,
		}
		storemgr = storemanager.NewStoreManager(tables, storeCfg, fs)
		storemgr.Start()
	}

	displayContext := kubeContext
	if conf.DisplayContext != "" {
		displayContext = conf.DisplayContext
	}

	webConfig := webserver.WebConfig{
		BindAddress:      conf.BindAddress,
		Port:             conf.Port,
		WebFilesPath:     conf.WebFilesPath,
		ConfigYaml:       conf.ToYaml(),
		MaxLookback:      conf.MaxLookback,
		DefaultNamespace: conf.DefaultNamespace,
		DefaultLookback:  conf.DefaultLookback,
		DefaultResources: conf.DefaultKind,
		ResourceLinks:    conf.ResourceLinks,
		LeftBarLinks:     conf.LeftBarLinks,
		CurrentContext:   displayContext,
	}
	err = webserver.Run(webConfig, tables)
	if err != nil {
		return errors.Wrap(err, "failed to run webserver")
	}

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
