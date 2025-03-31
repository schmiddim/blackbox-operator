package controller

import (
	"context"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"github.com/schmiddim/blackbox-operator/pkg/monitoring"
	istioNetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ServiceEntryReconciler reconciles a ServiceEntry object
type ServiceEntryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *config.Config
}

// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=create;list;get;update;patch;delete;watch
// +kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries,verbs=create;list;get;watch
// +kubebuilder:rbac:groups=monitoring.coreos.com/v1,resources=servicemonitors,verbs=create;list;get;update;patch;delete;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceEntry object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *ServiceEntryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	smm := monitoring.NewServiceMonitorMapper(r.Config, &logger)
	exclude := monitoring.NewExcluded(r.Config)
	// Try to fetch the ServiceEntry
	var se istioNetworking.ServiceEntry
	if err := r.Get(ctx, req.NamespacedName, &se); err != nil {
		if errors.IsNotFound(err) {
			// ServiceEntry was deleted → Delete the associated ServiceMonitor
			smName, err := smm.GetNameForServiceMonitor(req.Name)
			if err != nil {
				return ctrl.Result{}, err
			}
			sm := &monitoringv1.ServiceMonitor{}
			err = r.Get(ctx, client.ObjectKey{Name: smName, Namespace: req.Namespace}, sm)
			if err == nil {
				// Delete the ServiceMonitor if it exists
				err = r.Delete(ctx, sm)
				if err != nil {
					return ctrl.Result{}, err
				}
				logger.Info("ServiceMonitor deleted", "name", smName)
			} else if !errors.IsNotFound(err) {
				// Return error if it's not a "not found" error
				return ctrl.Result{}, err
			}
			// No further action required
			return ctrl.Result{}, nil
		}
		// Return any other error
		return ctrl.Result{}, err
	}

	logger.Info("ServiceEntry detected/modified", "name", se.Name, "namespace", se.Namespace)

	// Generate the desired ServiceMonitor based on the ServiceEntry
	sm := smm.MapperForService(&se)

	existingSM := &monitoringv1.ServiceMonitor{}
	err := r.Get(ctx, client.ObjectKey{Name: sm.Name, Namespace: sm.Namespace}, existingSM)

	if exclude.IsExcluded(se.ObjectMeta.Labels) {
		logger.Info("No ServiceMonitor created because of ExcludeRules", "name", se.Name, "namespace", se.Namespace)
		if err == nil && errors.IsNotFound(err) == false {
			logger.Info("Service Monitor exists and will be deleted", "name", existingSM.Name, "namespace", existingSM.Namespace)
			err := r.Delete(ctx, existingSM)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err != nil && errors.IsNotFound(err) && exclude.IsExcluded(se.ObjectMeta.Labels) == false {
		// ServiceMonitor does not exist Create it
		err = r.Create(ctx, sm)
		if err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("ServiceMonitor created", "name", sm.Name)
	} else if err == nil {
		// Compare existing ServiceMonitor with desired state to avoid unnecessary updates
		if !serviceMonitorEqual(existingSM, sm) {
			patch := client.MergeFrom(existingSM.DeepCopy())
			existingSM.Spec = sm.Spec
			existingSM.Labels = sm.Labels
			err = r.Patch(ctx, existingSM, patch)
			if err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("ServiceMonitor updated", "name", sm.Name)
		} else {
			logger.Info("ServiceMonitor unchanged", "name", sm.Name)
		}
	} else {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func serviceMonitorEqual(a, b *monitoringv1.ServiceMonitor) bool {
	return reflect.DeepEqual(a.Spec, b.Spec) && reflect.DeepEqual(a.Labels, b.Labels)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&istioNetworking.ServiceEntry{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		Complete(r)
}
