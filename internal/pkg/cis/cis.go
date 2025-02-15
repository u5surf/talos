/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cis

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/talos-systems/talos/pkg/constants"
)

// const disabled = "false"

const auditPolicy string = `apiVersion: audit.k8s.io/v1beta1
kind: Policy
rules:
- level: Metadata
`

const encryptionConfig string = `kind: EncryptionConfig
apiVersion: v1
resources:
- resources:
  - secrets
  providers:
  - aescbc:
      keys:
      - name: key1
        secret: {{ .AESCBCEncryptionSecret }}
  - identity: {}
`

// EnforceAuditingRequirements enforces CIS requirements for auditing.
func EnforceAuditingRequirements() error {
	// // TODO(andrewrynhard): We should log to a file, and the option to retrieve
	// // the log files.
	// cfg.APIServer.ExtraArgs["audit-log-path"] = "-"
	// cfg.APIServer.ExtraArgs["audit-log-maxage"] = "30"
	// cfg.APIServer.ExtraArgs["audit-log-maxbackup"] = "3"
	// cfg.APIServer.ExtraArgs["audit-log-maxsize"] = "50"

	return nil
}

// WriteAuditPolicyToDisk writes the audit policy to disk.
func WriteAuditPolicyToDisk() (err error) {
	return ioutil.WriteFile(constants.AuditPolicyPath, []byte(auditPolicy), 0400)
}

// CreateEncryptionToken generates an encryption token to be used for secrets.
func CreateEncryptionToken() (string, error) {
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return "", err
	}

	str := base64.StdEncoding.EncodeToString(encryptionKey)

	return str, nil
}

// WriteEncryptionConfigToDisk writes an EncryptionConfig to disk.
func WriteEncryptionConfigToDisk(aescbcEncryptionSecret string) error {
	if _, err := os.Stat(constants.EncryptionConfigPath); os.IsNotExist(err) {
		aux := struct {
			AESCBCEncryptionSecret string
		}{
			AESCBCEncryptionSecret: aescbcEncryptionSecret,
		}
		t, err := template.New("encryptionconfig").Parse(encryptionConfig)
		if err != nil {
			return err
		}

		encBytes := []byte{}
		buf := bytes.NewBuffer(encBytes)
		if err := t.Execute(buf, aux); err != nil {
			return err
		}
		if err := ioutil.WriteFile(constants.EncryptionConfigPath, buf.Bytes(), 0400); err != nil {
			return err
		}
	}

	return nil
}

// EnforceSecretRequirements enforces CIS requirements for secrets.
func EnforceSecretRequirements() error {
	// cfg.APIServer.ExtraArgs["experimental-encryption-provider-config"] = constants.EncryptionConfigRootfsPath
	// vol := kubeadmapi.HostPathMount{
	// 	Name:      "encryptionconfig",
	// 	HostPath:  constants.EncryptionConfigRootfsPath,
	// 	MountPath: constants.EncryptionConfigRootfsPath,
	// 	ReadOnly:  true,
	// 	PathType:  v1.HostPathFile,
	// }
	// cfg.APIServer.ExtraVolumes = append(cfg.APIServer.ExtraVolumes, vol)

	return nil
}

// EnforceTLSRequirements enforces CIS requirements for TLS.
func EnforceTLSRequirements() error {
	// // nolint: lll
	// cfg.APIServer.ExtraArgs["tls-cipher-suites"] = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256"

	return nil
}

// EnforceAdmissionPluginsRequirements enforces CIS requirements for admission plugins.
// TODO(andrewrynhard): Include any extra user specified plugins.
// TODO(andrewrynhard): Enable EventRateLimit.
// TODO(andrewrynhard): Enable AlwaysPullImages (See https://github.com/kubernetes/kubernetes/issues/64333).
func EnforceAdmissionPluginsRequirements() error {
	// // nolint: lll
	// cfg.APIServer.ExtraArgs["enable-admission-plugins"] = "PodSecurityPolicy,NamespaceLifecycle,ServiceAccount,NodeRestriction,LimitRanger,DefaultStorageClass,DefaultTolerationSeconds,ResourceQuota"

	return nil
}

// EnforceExtraRequirements enforces miscellaneous CIS requirements.
// TODO(andrewrynhard): Enable anonymous-auth, see https://github.com/kubernetes/kubeadm/issues/798.
// TODO(andrewrynhard): Enable kubelet-certificate-authority, see https://github.com/kubernetes/kubeadm/issues/118#issuecomment-407202481.
func EnforceExtraRequirements() error {
	// cfg.APIServer.ExtraArgs["profiling"] = disabled
	// cfg.ControllerManager.ExtraArgs["profiling"] = disabled
	// cfg.Scheduler.ExtraArgs["profiling"] = disabled

	// cfg.APIServer.ExtraArgs["service-account-lookup"] = "true"

	return nil
}

// EnforceBootstrapMasterRequirements enforces the CIS requirements for master nodes.
func EnforceBootstrapMasterRequirements() error {
	// ensureFieldsAreNotNil(cfg)

	// if err := EnforceAuditingRequirements(cfg); err != nil {
	// 	return err
	// }

	// if err := EnforceSecretRequirements(cfg); err != nil {
	// 	return err
	// }

	// if err := EnforceTLSRequirements(cfg); err != nil {
	// 	return err
	// }

	// if err := EnforceAdmissionPluginsRequirements(cfg); err != nil {
	// 	return err
	// }

	// if err := EnforceExtraRequirements(cfg); err != nil {
	// 	return err
	// }

	return nil
}

// EnforceCommonMasterRequirements enforces the CIS requirements for master nodes.
func EnforceCommonMasterRequirements(aescbcEncryptionSecret string) (err error) {
	if err = WriteAuditPolicyToDisk(); err != nil {
		return err
	}

	if err = WriteEncryptionConfigToDisk(aescbcEncryptionSecret); err != nil {
		return err
	}

	return nil
}

// EnforceWorkerRequirements enforces the CIS requirements for master nodes.
func EnforceWorkerRequirements() error {
	return nil
}
