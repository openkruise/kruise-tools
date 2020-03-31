/*
Copyright 2020 The Kruise Authors.

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

package migrate

import (
	"fmt"
	"time"

	cmdutil "github.com/openkruise/kruise-tools/cmd/util"
	"github.com/openkruise/kruise-tools/pkg/creation"
	clonesetcreation "github.com/openkruise/kruise-tools/pkg/creation/cloneset"
	"github.com/openkruise/kruise-tools/pkg/migration"
	clonesetmigration "github.com/openkruise/kruise-tools/pkg/migration/cloneset"
	"github.com/spf13/cobra"
)

func (o *migrateOptions) migrateCloneSet(f cmdutil.Factory, cmd *cobra.Command) error {
	cfg, err := f.ToRESTConfig()
	if err != nil {
		return err
	}

	if o.IsCreate {

		ctrl, err := clonesetcreation.NewControl(cfg)
		if err != nil {
			return err
		}

		opts := creation.Options{CopyReplicas: o.IsCopy}
		if err := ctrl.Create(o.SrcRef, o.DstRef, opts); err != nil {
			return err
		}

		cmdutil.Print(fmt.Sprintf("Successfully created from %s/%s to %s/%s", o.From, o.SrcName, o.To, o.DstName))

	} else {

		stopChan := make(chan struct{})
		ctrl, err := clonesetmigration.NewControl(cfg, stopChan)
		if err != nil {
			return err
		}

		opts := migration.Options{}
		if o.Replicas >= 0 {
			opts.Replicas = &o.Replicas
		}
		if o.MaxSurge >= 1 {
			opts.MaxSurge = &o.MaxSurge
		}
		if o.TimeoutSeconds > 0 {
			opts.TimeoutSeconds = &o.TimeoutSeconds
		}

		oldResult, err := ctrl.Submit(o.SrcRef, o.DstRef, opts)
		if err != nil {
			return err
		}

		for {
			time.Sleep(time.Second)
			newResult, err := ctrl.Query(oldResult.ID)
			if err != nil {
				return err
			}

			if newResult.SrcMigratedReplicas != oldResult.SrcMigratedReplicas || newResult.DstMigratedReplicas != oldResult.DstMigratedReplicas {
				cmdutil.Print(fmt.Sprintf("Migration progress: %s/%s scale in %d, %s/%s scale out %d",
					o.From, o.SrcName, newResult.SrcMigratedReplicas, o.To, o.DstName, newResult.DstMigratedReplicas))
			}

			switch newResult.State {
			case migration.MigrateSucceeded:
				cmdutil.Print(fmt.Sprintf("Successfully migrated %v replicas from %s/%s to %s/%s",
					newResult.DstMigratedReplicas, o.From, o.SrcName, o.To, o.DstName))
				return nil
			case migration.MigrateFailed:
				return fmt.Errorf("failed to migrate: %v", newResult.Message)
			}

			oldResult = newResult
		}
	}

	return nil
}
