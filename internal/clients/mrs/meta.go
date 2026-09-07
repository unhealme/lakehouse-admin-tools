package mrs

import (
	"net/http"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
)

func GenReqDefForListClusterManagerAuthState() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/clusters/{cluster_id}/manager-token").
		WithResponse(new(GetClusterManagerTokenResponse)).
		WithContentType("application/json")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ClusterId").
		WithJsonTag("cluster_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}
