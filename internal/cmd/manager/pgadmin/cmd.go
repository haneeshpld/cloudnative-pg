/*
Copyright © contributors to CloudNativePG, established as
CloudNativePG a Series of LF Projects, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

// Package pgadmin implements the "pgadmin" command to start the PgAdmin controller manager.
package pgadmin

import (
	"github.com/cloudnative-pg/machinery/pkg/log"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	postgresqlv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	"github.com/cloudnative-pg/cloudnative-pg/internal/controller"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	// Register core Kubernetes types
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// Register CloudNativePG PgAdmin CRD types
	utilruntime.Must(postgresqlv1.AddToScheme(scheme))
}

// NewCmd creates a new cobra command to run the PgAdmin controller manager.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pgadmin",
		Short: "Start the PgAdmin controller manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.FromContext(cmd.Context())

			// Get a config to talk to the apiserver
			cfg := config.GetConfigOrDie()

			// Create a new manager to provide shared dependencies and start components
			mgr, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: scheme,
			})
			if err != nil {
				logger.Error(err, "unable to start manager")
				return err
			}

			// Setup the PgAdmin controller with the Manager.
			if err = (&controller.PgAdminReconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				logger.Error(err, "unable to create PgAdmin controller")
				return err
			}

			logger.Info("starting manager")
			if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
				logger.Error(err, "problem running manager")
				return err
			}

			return nil
		},
	}

	return cmd
}
