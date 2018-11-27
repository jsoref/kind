/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"testing"

	"sigs.k8s.io/kind/pkg/util"
)

// TODO(fabriziopandini): ideally this should use scheme.Default, but this creates a circular dependency
// So the current solution is to mimic defaulting for the validation test, but probably there should be a better solution here
func newDefaultedConfig() *Config {
	cfg := &Config{
		Image: "myImage:latest",
	}
	return cfg
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		TestName       string
		Config         *Config
		ExpectedErrors int
	}{
		{
			TestName:       "Canonical config",
			Config:         newDefaultedConfig(),
			ExpectedErrors: 0,
		},
		{
			TestName: "Invalid PreBoot hook",
			Config: func() *Config {
				cfg := newDefaultedConfig()
				cfg.ControlPlane = &ControlPlane{
					NodeLifecycle: &NodeLifecycle{
						PreBoot: []LifecycleHook{
							{
								Command: []string{},
							},
						},
					},
				}
				return cfg
			}(),
			ExpectedErrors: 1,
		},
		{
			TestName: "Invalid PreKubeadm hook",
			Config: func() *Config {
				cfg := newDefaultedConfig()
				cfg.ControlPlane = &ControlPlane{
					NodeLifecycle: &NodeLifecycle{
						PreKubeadm: []LifecycleHook{
							{
								Name:    "pull an image",
								Command: []string{},
							},
						},
					},
				}
				return cfg
			}(),
			ExpectedErrors: 1,
		},
		{
			TestName: "Invalid PostKubeadm hook",
			Config: func() *Config {
				cfg := newDefaultedConfig()
				cfg.ControlPlane = &ControlPlane{
					NodeLifecycle: &NodeLifecycle{
						PostKubeadm: []LifecycleHook{
							{
								Name:    "pull an image",
								Command: []string{},
							},
						},
					},
				}
				return cfg
			}(),
			ExpectedErrors: 1,
		},
		{
			TestName: "Empty image field",
			Config: func() *Config {
				cfg := newDefaultedConfig()
				cfg.Image = ""
				return cfg
			}(),
			ExpectedErrors: 1,
		},
	}

	for _, tc := range cases {
		err := tc.Config.Validate()
		// the error can be:
		// - nil, in which case we should expect no errors or fail
		if err == nil {
			if tc.ExpectedErrors != 0 {
				t.Errorf("received no errors but expected errors for case %s", tc.TestName)
			}
			continue
		}
		// - not castable to *Errors, in which case we have the wrong error type ...
		configErrors, ok := err.(*util.Errors)
		if !ok {
			t.Errorf("config.Validate should only return nil or ConfigErrors{...}, got: %v for case: %s", err, tc.TestName)
			continue
		}
		// - ConfigErrors, in which case expect a certain number of errors
		errors := configErrors.Errors()
		if len(errors) != tc.ExpectedErrors {
			t.Errorf("expected %d errors but got len(%v) = %d for case: %s", tc.ExpectedErrors, errors, len(errors), tc.TestName)
		}
	}
}
