package migrate

import (
	"fmt"
	"time"

	internalcmdutil "github.com/openkruise/kruise-tools/pkg/cmd/util"
	daemonsetcreation "github.com/openkruise/kruise-tools/pkg/creation/daemonset"
	"github.com/openkruise/kruise-tools/pkg/migration"
	daemonsetmigration "github.com/openkruise/kruise-tools/pkg/migration/daemonset"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// migrateDaemonSet handles both --create/--copy and full migration.
func (o *migrateOptions) migrateDaemonSet(f cmdutil.Factory, cmd *cobra.Command) error {
	cfg, err := f.ToRESTConfig()
	if err != nil {
		return err
	}

	// CREATE path
	if o.IsCreate {
		ctrl, err := daemonsetcreation.NewControl(cfg)
		if err != nil {
			return err
		}
		opts := daemonsetcreation.Options{CopyReplicas: o.IsCopy}
		if err := ctrl.Create(o.SrcRef, o.DstRef, opts); err != nil {
			return err
		}
		internalcmdutil.Print(fmt.Sprintf(
			"Successfully created AdvancedDaemonSet %s/%s from DaemonSet %s/%s",
			o.Namespace, o.DstName, o.Namespace, o.SrcName))
		return nil
	}

	// MIGRATE path
	stopChan := make(chan struct{})
	ctrl, err := daemonsetmigration.NewControl(cfg, stopChan)
	if err != nil {
		return err
	}

	result, err := ctrl.Submit(o.SrcRef, o.DstRef, migration.Options{})
	if err != nil {
		return err
	}

	startTime := time.Now()
	for {
		if o.TimeoutSeconds > 0 && time.Since(startTime).Seconds() > float64(o.TimeoutSeconds) {
			return fmt.Errorf("migration timed out after %d seconds", o.TimeoutSeconds)
		}
		time.Sleep(time.Second)
		newResult, err := ctrl.Query(result.ID)
		if err != nil {
			return err
		}
		if newResult.State == migration.MigrateSucceeded {
			internalcmdutil.Print(fmt.Sprintf(
				"Successfully migrated DaemonSet %s/%s to AdvancedDaemonSet %s/%s",
				o.Namespace, o.SrcName, o.Namespace, o.DstName))
			return nil
		}
		if newResult.State == migration.MigrateFailed {
			return fmt.Errorf("migration failed: %s", newResult.Message)
		}
	}
}
