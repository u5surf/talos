/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package kspp

import (
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/talos-systems/talos/internal/pkg/kernel"
	"github.com/talos-systems/talos/pkg/sysctl"
)

// RequiredKSPPKernelParameters is the set of kernel parameters required to
// satisfy the KSPP.
var RequiredKSPPKernelParameters = kernel.Parameters{
	kernel.NewParameter("page_poison").Append("1"),
	kernel.NewParameter("slab_nomerge").Append(""),
	kernel.NewParameter("slub_debug").Append("P"),
	kernel.NewParameter("pti").Append("on"),
}

// EnforceKSPPKernelParameters verifies that all required KSPP kernel
// parameters are present with the right value.
func EnforceKSPPKernelParameters() error {
	var result *multierror.Error
	for _, values := range RequiredKSPPKernelParameters {
		var val *string
		if val = kernel.ProcCmdline().Get(values.Key()).First(); val == nil {
			result = multierror.Append(result, errors.Errorf("KSPP kernel parameter %s is required", values.Key()))
			continue
		}

		expected := values.First()
		if *val != *expected {
			result = multierror.Append(result, errors.Errorf("KSPP kernel parameter %s was found with value %s, expected %s", values.Key(), *val, *expected))
		}
	}

	return result.ErrorOrNil()
}

// EnforceKSPPSysctls verifies that all required KSPP kernel sysctls are set
// with the right value.
func EnforceKSPPSysctls() (err error) {
	props := []*sysctl.SystemProperty{
		{
			Key:   "kernel.kptr_restrict",
			Value: "1",
		},
		{
			Key:   "kernel.dmesg_restrict",
			Value: "1",
		},
		{
			Key:   "kernel.perf_event_paranoid",
			Value: "3",
		},
		// We can skip this sysctl because CONFIG_KEXEC is not set.
		// {
		// 	Key:   "kernel.kexec_load_disabled",
		// 	Value: "1",
		// },
		{
			Key:   "kernel.yama.ptrace_scope",
			Value: "1",
		},
		{
			Key:   "user.max_user_namespaces",
			Value: "0",
		},
		{
			Key:   "kernel.unprivileged_bpf_disabled",
			Value: "1",
		},
		{
			Key:   "net.core.bpf_jit_harden",
			Value: "2",
		},
	}

	for _, prop := range props {
		if err = sysctl.WriteSystemProperty(prop); err != nil {
			return
		}
	}

	return nil
}
