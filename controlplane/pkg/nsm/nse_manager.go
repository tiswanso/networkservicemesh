package nsm

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"

	"github.com/networkservicemesh/networkservicemesh/sdk/common"

	"github.com/sirupsen/logrus"

	local "github.com/networkservicemesh/networkservicemesh/controlplane/pkg/apis/local/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/apis/nsm"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/apis/nsm/connection"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/apis/registry"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/model"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/serviceregistry"
)

type networkServiceEndpointManager interface {
	getEndpoint(ctx context.Context, requestConnection connection.Connection, ignoreEndpoints map[string]*registry.NSERegistration) (*registry.NSERegistration, error)
	createNSEClient(ctx context.Context, endpoint *registry.NSERegistration) (nsm.NetworkServiceClient, error)
	isLocalEndpoint(endpoint *registry.NSERegistration) bool
	checkUpdateNSE(ctx context.Context, reg *registry.NSERegistration) bool
}

type nseManager struct {
	serviceRegistry serviceregistry.ServiceRegistry
	model           model.Model
	properties      *nsm.NsmProperties
}

func (nsem *nseManager) getEndpoint(ctx context.Context, requestConnection connection.Connection, ignoreEndpoints map[string]*registry.NSERegistration) (*registry.NSERegistration, error) {

	myNsemName := nsem.model.GetNsm().GetName()
	targetNsemName := requestConnection.GetDestinationNetworkServiceManagerName()
	// Handle case we are remote NSM and asked for particular endpoint to connect to.
	targetEndpoint := requestConnection.GetNetworkServiceEndpointName()
	if len(targetEndpoint) > 0 {
		if len(targetNsemName) > 0 && myNsemName == targetNsemName {
			endpoint := nsem.model.GetEndpoint(targetEndpoint)
			if endpoint != nil && ignoreEndpoints[endpoint.EndpointName()] == nil {
				return endpoint.Endpoint, nil
			} else {
				return nil, fmt.Errorf("Could not find endpoint with name: %s at local registry", targetEndpoint)
			}
		}
	}

	// Get endpoints, do it every time since we do not know if list are changed or not.
	discoveryClient, err := nsem.serviceRegistry.DiscoveryClient()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	nseRequest := &registry.FindNetworkServiceRequest{
		NetworkServiceName: requestConnection.GetNetworkService(),
	}
	endpointResponse, err := discoveryClient.FindNetworkService(ctx, nseRequest)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	var endpoint *registry.NetworkServiceEndpoint
	if len(targetEndpoint) > 0 {
		// Connection is to specific target endpoint
		endpoint, err = nsem.matchTargetEndpoint(requestConnection, endpointResponse.GetNetworkServiceEndpoints())
	}
	if endpoint == nil {
		endpoints := nsem.filterEndpoints(endpointResponse.GetNetworkServiceEndpoints(), ignoreEndpoints)

		if len(endpoints) == 0 {
			return nil, fmt.Errorf("failed to find NSE for NetworkService %s. Checked: %d of total NSEs: %d",
				requestConnection.GetNetworkService(), len(ignoreEndpoints), len(endpoints))
		}

		endpoint = nsem.model.GetSelector().SelectEndpoint(requestConnection.(*local.Connection), endpointResponse.GetNetworkService(), endpoints)
		if endpoint == nil {
			return nil, fmt.Errorf("failed to find NSE for NetworkService %s. Checked: %d of total NSEs: %d",
				requestConnection.GetNetworkService(), len(ignoreEndpoints), len(endpoints))
		}
	}

	return &registry.NSERegistration{
		NetworkServiceManager:  endpointResponse.GetNetworkServiceManagers()[endpoint.GetNetworkServiceManagerName()],
		NetworkServiceEndpoint: endpoint,
		NetworkService:         endpointResponse.GetNetworkService(),
	}, nil
}

func (nsem *nseManager) matchTargetEndpoint(requestConnection connection.Connection, endpoints []*registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	logrus.Infof("Matching target endpoint %s", requestConnection.GetNetworkServiceEndpointName())
	for _, endpoint := range endpoints {
		if endpoint.GetName() == requestConnection.GetNetworkServiceEndpointName() {
			logrus.Infof("Found target endpoint %s", requestConnection.GetNetworkServiceEndpointName())
			return endpoint, nil
		}
	}
	return nil, nil
}

/**
ctx - we assume it is big enought to perform connection.
*/
func (nsem *nseManager) createNSEClient(ctx context.Context, endpoint *registry.NSERegistration) (nsm.NetworkServiceClient, error) {

	var span opentracing.Span
	if opentracing.GlobalTracer() != nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, "nsm.create.nse.client")
		defer span.Finish()
	}

	logger := common.LogFromSpan(span)
	if nsem.isLocalEndpoint(endpoint) {
		modelEp := nsem.model.GetEndpoint(endpoint.GetNetworkServiceEndpoint().GetName())
		if modelEp == nil {
			return nil, fmt.Errorf("Endpoint not found: %v", endpoint)
		}
		logger.Infof("Create local NSE connection to endpoint: %v", modelEp)
		client, conn, err := nsem.serviceRegistry.EndpointConnection(ctx, modelEp)
		if err != nil {
			// We failed to connect to local NSE.
			nsem.cleanupNSE(modelEp)
			return nil, err
		}
		return &endpointClient{connection: conn, client: client}, nil
	} else {
		logger.Infof("Create remote NSE connection to endpoint: %v", endpoint)
		client, conn, err := nsem.serviceRegistry.RemoteNetworkServiceClient(ctx, endpoint.GetNetworkServiceManager())
		if err != nil {
			return nil, err
		}
		return &nsmClient{client: client, connection: conn}, nil
	}
}

func (nsem *nseManager) isLocalEndpoint(endpoint *registry.NSERegistration) bool {
	return nsem.model.GetNsm().GetName() == endpoint.GetNetworkServiceEndpoint().GetNetworkServiceManagerName()
}

func (nsem *nseManager) checkUpdateNSE(ctx context.Context, reg *registry.NSERegistration) bool {
	pingCtx, pingCancel := context.WithTimeout(ctx, nsem.properties.HealRequestConnectCheckTimeout)
	defer pingCancel()

	client, err := nsem.createNSEClient(pingCtx, reg)
	if err == nil && client != nil {
		_ = client.Cleanup()
		return true
	}

	return false
}

func (nsem *nseManager) cleanupNSE(endpoint *model.Endpoint) {
	// Remove endpoint from model and put workspace into BAD state.
	nsem.model.DeleteEndpoint(endpoint.EndpointName())
	logrus.Infof("NSM: Remove Endpoint since it is not available... %v", endpoint)
}

func (nsem *nseManager) filterEndpoints(endpoints []*registry.NetworkServiceEndpoint, ignoreEndpoints map[string]*registry.NSERegistration) []*registry.NetworkServiceEndpoint {
	result := []*registry.NetworkServiceEndpoint{}
	// Do filter of endpoints
	for _, candidate := range endpoints {
		if ignoreEndpoints[candidate.GetName()] == nil {
			result = append(result, candidate)
		}
	}
	return result
}
