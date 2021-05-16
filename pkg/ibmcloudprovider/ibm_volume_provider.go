/**
 * Copyright 2020 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package ibmcloudprovider ...
package ibmcloudprovider

import (
	"github.com/IBM/ibmcloud-volume-interface/config"
	"github.com/IBM/ibmcloud-volume-interface/lib/provider"
	"github.com/IBM/ibm-csi-common/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// CloudProviderInterface ...
type CloudProviderInterface interface {
	GetProviderSession(ctx context.Context, logger *zap.Logger) (provider.Session, error)
	GetConfig() *config.Config
	GetClusterInfo() *utils.ClusterInfo
}
