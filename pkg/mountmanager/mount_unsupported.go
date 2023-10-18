//go:build !linux && !windows
// +build !linux,!windows

/**
 * Copyright 2021 IBM Corp.
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

// Package mountmanager ...
package mountmanager

import (
	"errors"

	mount "k8s.io/mount-utils"
)

var errUnsupported = errors.New("util/mount on this platform is not supported")

// MakeFile ...
func (m *NodeMounter) MakeFile(pathname string) error {
	return errUnsupported
}

// MakeDir ...
func (m *NodeMounter) MakeDir(pathname string) error {
	return errUnsupported
}

// PathExists ...
func (m *NodeMounter) PathExists(pathname string) (bool, error) {
	return true, errors.New("not implemented")
}

// GetSafeFormatAndMount returns the existing SafeFormatAndMount object of NodeMounter.
func (m *NodeMounter) GetSafeFormatAndMount() *mount.SafeFormatAndMount {
	return nil
}

// Resize returns boolean and error if any
func (m *NodeMounter) Resize(devicePath string, deviceMountPath string) (bool, error) {
	return true, errors.New("not implemented")
}

// MountEITBasedFileShare ...
func (m *NodeMounter) MountEITBasedFileShare(stagingTargetPath string, targetPath string, fsType string, requestID string) error {
	return nil
}

// UmountEITBasedFileShare ...
func (m *NodeMounter) UmountEITBasedFileShare(targetPath string, requestID string) error {
	return nil
}

// DebugLogsEITBasedFileShare...
func (m *NodeMounter) DebugLogsEITBasedFileShare(requestID string) error {
	return nil
}
