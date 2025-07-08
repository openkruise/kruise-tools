package migrate

import (
	"fmt"
	"time"

	internalcmdutil "github.com/openkruise/kruise-tools/pkg/cmd/util"
	sscreation "github.com/openkruise/kruise-tools/pkg/creation/statefulset"
	"github.com/openkruise/kruise-tools/pkg/migration"
	ssmigration "github.com/openkruise/kruise-tools/pkg/migration/statefulset"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// migrateStatefulSet handles both --create/--copy and full migration.
func (o *migrateOptions) migrateStatefulSet(f cmdutil.Factory, cmd *cobra.Command) error {
	cfg, err := f.ToRESTConfig()
	if err != nil {
		return err
	}

	// CREATE path
	if o.IsCreate {
		ctrl, err := sscreation.NewControl(cfg)
		if err != nil {
			return err
		}
		opts := sscreation.Options{CopyReplicas: o.IsCopy}
		if err := ctrl.Create(o.SrcRef, o.DstRef, opts); err != nil {
			return err
		}
		internalcmdutil.Print(fmt.Sprintf(
			"Successfully created AdvancedStatefulSet %s/%s from StatefulSet %s/%s",
			o.Namespace, o.DstName, o.Namespace, o.SrcName))
		return nil
	}

	// MIGRATE path
	stopChan := make(chan struct{})
	ctrl, err := ssmigration.NewControl(cfg, stopChan)
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
				"Successfully migrated StatefulSet %s/%s to AdvancedStatefulSet %s/%s",
				o.Namespace, o.SrcName, o.Namespace, o.DstName))
			return nil
		}
		if newResult.State == migration.MigrateFailed {
			return fmt.Errorf("migration failed: %s", newResult.Message)
		}
	}
}
