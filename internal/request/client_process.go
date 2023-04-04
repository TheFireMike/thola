//go:build client
// +build client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/doc"
	"github.com/inexio/thola/internal/network"
	"github.com/inexio/thola/internal/parser"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"strings"
)

func (r *IdentifyRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "identify", apiFormat)
	if err != nil {
		return nil, err
	}
	var res IdentifyResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *CheckIdentifyRequest) process(ctx context.Context) (Response, error) {
	var res CheckIdentifyResponse
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "check/identify", apiFormat)
	if err != nil {
		m := monitoringplugin.NewResponse("")
		m.UpdateStatusOnError(err, 3, "failed to send request to api", true)
		res.ResponseInfo = m.GetInfo()
		return &res, nil
	}
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		m := monitoringplugin.NewResponse("")
		m.UpdateStatusOnError(err, 3, "failed to parse response from thola api to icinga output", true)
		res.ResponseInfo = m.GetInfo()
	}
	return &res, nil
}

func (r *CheckSNMPRequest) process(ctx context.Context) (Response, error) {
	var res CheckSNMPResponse
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "check/snmp", apiFormat)
	if err != nil {
		m := monitoringplugin.NewResponse("")
		m.UpdateStatusOnError(err, 3, "failed to send request to api", true)
		res.ResponseInfo = m.GetInfo()
		return &res, nil
	}
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		m := monitoringplugin.NewResponse("")
		m.UpdateStatusOnError(err, 3, "failed to parse response from thola api to icinga output", true)
		res.ResponseInfo = m.GetInfo()
	}
	return &res, nil
}

func (r *CheckInterfaceMetricsRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/interface-metrics"), nil
}

func (r *CheckTholaServerRequest) process(ctx context.Context) (Response, error) {
	var res CheckResponse
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "check/thola-server", apiFormat)
	if err != nil {
		m := monitoringplugin.NewResponse("")
		if tholaerr.IsNetworkError(err) {
			m.UpdateStatusOnError(err, 2, "no thola server found", true)
		} else {
			m.UpdateStatusOnError(err, 2, "failed to query thola server", true)
		}

		res.ResponseInfo = m.GetInfo()
		return &res, nil
	}
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		m := monitoringplugin.NewResponse("")
		m.UpdateStatusOnError(err, 3, "failed to parse response", true)
		res.ResponseInfo = m.GetInfo()
	}
	return &res, nil
}

func (r *CheckUPSRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/ups"), nil
}

func (r *CheckSBCRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/sbc"), nil
}

func (r *CheckMemoryUsageRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/memory-usage"), nil
}

func (r *CheckServerRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/server"), nil
}

func (r *CheckDiskRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/disk"), nil
}

func (r *CheckCPULoadRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/cpu-load"), nil
}

func (r *CheckHardwareHealthRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/hardware-health"), nil
}

func (r *CheckHighAvailabilityRequest) process(ctx context.Context) (Response, error) {
	return checkProcess(ctx, r, "check/high-availability"), nil
}

func (r *ReadInterfacesRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/interfaces", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadInterfacesResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadCountInterfacesRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/count-interfaces", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadCountInterfacesResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadCPULoadRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/cpu-load", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadCPULoadResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadMemoryUsageRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/memory-usage", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadMemoryUsageResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadUPSRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/ups", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadUPSResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadSBCRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/sbc", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadSBCResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadServerRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/server", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadServerResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadDiskRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/disk", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadDiskResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadHardwareHealthRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/hardware-health", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadHardwareHealthResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadHighAvailabilityRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/high-availability", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadHighAvailabilityResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func (r *ReadAvailableComponentsRequest) process(ctx context.Context) (Response, error) {
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, "read/available-components", apiFormat)
	if err != nil {
		return nil, err
	}
	var res ReadAvailableComponentsResponse
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse api response body to thola response")
	}
	return &res, nil
}

func checkProcess(ctx context.Context, r Request, apiPath string) Response {
	var res CheckResponse
	apiFormat := viper.GetString("target-api-format")
	responseBody, err := sendToAPI(ctx, r, apiPath, apiFormat)
	if err != nil {
		m := monitoringplugin.NewResponse("")
		m.UpdateStatusOnError(err, 3, "failed to send request to api", true)
		res.ResponseInfo = m.GetInfo()
		return &res
	}
	err = parser.ToStruct(responseBody, apiFormat, &res)
	if err != nil {
		m := monitoringplugin.NewResponse("")
		m.UpdateStatusOnError(err, 3, "failed to parse response from thola api to icinga output", true)
		res.ResponseInfo = m.GetInfo()
	}
	return &res
}

func sendToAPI(ctx context.Context, request Request, path, format string) ([]byte, error) {
	apiUserName := viper.GetString("target-api-username")
	apiPassword := viper.GetString("target-api-password")

	client, err := network.NewHTTPClient(viper.GetString("target-api"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http client")
	}

	if apiUserName != "" && apiPassword != "" {
		err := client.SetUsernameAndPassword(apiUserName, apiPassword)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set api username and password")
		}
	}

	client.InsecureSSLCert(viper.GetBool("insecure-ssl-cert"))
	err = client.SetFormat(viper.GetString("target-api-format"))
	if err != nil {
		return nil, errors.Wrap(err, "error during set format of http client")
	}

	b, err := parser.Parse(request, format)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse request to format '%s'", format)
	}

	header := map[string]string{"User-Agent": "Thola Client " + doc.Version}
	rid, ok := RequestIDFromContext(ctx)
	if ok {
		header["X-Request-ID"] = rid
	}

	restyResponse, err := client.Request(ctx, "POST", path, string(b), header, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request to api")
	}

	if restyResponse.IsError() {
		var errorMessageFetcher map[string]interface{}
		err = parser.ToStruct(restyResponse.Body(), format, &errorMessageFetcher)
		if err != nil {
			resStr := strings.Trim(fmt.Sprintf("%s", restyResponse.Body()), " \t\n")
			return nil, fmt.Errorf("an error occurred during api call. response body: '%s'", resStr)
		}
		if errMsg, ok := errorMessageFetcher["message"]; ok {
			return nil, fmt.Errorf("%s", errMsg)
		}
		if errMsg, ok := errorMessageFetcher["error"]; ok {
			return nil, fmt.Errorf("%s", errMsg)
		}
		return nil, fmt.Errorf("an error occurred during api call. response body: '%s'", restyResponse.Body())

	}

	return restyResponse.Body(), nil
}
