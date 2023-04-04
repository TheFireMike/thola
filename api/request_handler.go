package api

import (
	"context"
	"crypto/subtle"
	"fmt"
	"github.com/inexio/thola/api/statistics"
	"github.com/inexio/thola/internal/database"
	"github.com/inexio/thola/internal/request"
	"github.com/inexio/thola/internal/tholaerr"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

var deviceChannels struct {
	sync.RWMutex

	channels map[string]chan struct{}
}

// StartAPI starts the API.
func StartAPI() {
	ctx := log.Logger.WithContext(context.Background())

	log.Ctx(ctx).Debug().Msg("starting the server")

	db, err := database.GetDB(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("starting the server failed")
	}

	deviceChannels.channels = make(map[string]chan struct{})
	e := echo.New()

	e.HideBanner = true
	fmt.Print(" ______   __  __     ______     __         ______   \n" +
		"/\\__  _\\ /\\ \\_\\ \\   /\\  __ \\   /\\ \\       /\\  __ \\  \n" +
		"\\/_/\\ \\/ \\ \\  __ \\  \\ \\ \\/\\ \\  \\ \\ \\____  \\ \\  __ \\ \n" +
		"   \\ \\_\\  \\ \\_\\ \\_\\  \\ \\_____\\  \\ \\_____\\  \\ \\_\\ \\_\\\n" +
		"    \\/_/   \\/_/\\/_/   \\/_____/   \\/_____/   \\/_/\\/_/\n\n")

	if (viper.GetString("api.username") != "") && (viper.GetString("api.password") != "") {
		log.Ctx(ctx).Debug().Msg("set authorization for api")
		e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			// Be careful to use constant time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(username), []byte(viper.GetString("restapi.username"))) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(viper.GetString("restapi.password"))) == 1 {
				return true, nil
			}
			return false, nil
		}))
	}

	if viper.GetString("api.ratelimit") != "" {
		log.Ctx(ctx).Debug().Msg("set ratelimit for api")
		e.Use(ipRateLimit())
	}

	e.Use(statistics.Middleware())

	e.Use(requestIDMiddleware())

	e.Use(loggerMiddleware())

	// swagger:operation POST /identify identify identify
	// ---
	// summary: Identifies a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json<
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/IdentifyRequest'
	// responses:
	//   200:
	//     description: Returns the device which was found.
	//     schema:
	//       $ref: '#/definitions/IdentifyResponse'
	//   400:
	//     description: Returns a string that the request was formatted wrong.
	//   404:
	//     description: Returns a string that no device was found.
	e.POST("/identify", identify)

	// swagger:operation POST /check/identify check checkIdentify
	// ---
	// summary: Checks if identify matches the expectations.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckIdentifyRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckIdentifyResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/identify", checkIdentify)

	// swagger:operation POST /check/snmp check checkSNMP
	// ---
	// summary: Checks SNMP availability.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckSNMPRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckSNMPResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/snmp", checkSNMP)

	// swagger:operation POST /check/interface-metrics check checkInterfaceMetrics
	// ---
	// summary: Check to read out interface metrics.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckInterfaceMetricsRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/interface-metrics", checkInterfaceMetrics)

	// swagger:operation POST /check/thola-server check checkTholaServer
	// ---
	// summary: Check existence of thola servers.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckTholaServerRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/thola-server", checkTholaServer)

	// swagger:operation POST /check/ups check checkUPS
	// ---
	// summary: Checks whether a UPS device has its main voltage applied.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckUPSRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/ups", checkUPS)

	// swagger:operation POST /check/memory-usage check checkMemoryUsage
	// ---
	// summary: Check the memory usage of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckMemoryUsageRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/memory-usage", checkMemoryUsage)

	// swagger:operation POST /check/cpu-load check checkCPULoad
	// ---
	// summary: Check the cpu load of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckCPULoadRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/cpu-load", checkCPULoad)

	// swagger:operation POST /check/sbc check checkSBC
	// ---
	// summary: Check an sbc device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckSBCRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/sbc", checkSBC)

	// swagger:operation POST /check/server check checkServer
	// ---
	// summary: Check a linux server.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckServerRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/server", checkServer)

	// swagger:operation POST /check/disk check checkDisk
	// ---
	// summary: Check the disk of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckDiskRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/disk", checkDisk)

	// swagger:operation POST /check/hardware-health check checkHardwareHealth
	// ---
	// summary: Check the hardware health of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckHardwareHealthRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/hardware-health", checkHardwareHealth)

	// swagger:operation POST /check/high-availability check checkHighAvailability
	// ---
	// summary: Check the high availability status of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/CheckHighAvailabilityRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/CheckResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/check/high-availability", checkHighAvailability)

	// swagger:operation POST /read/interfaces read readInterfaces
	// ---
	// summary: Reads out data of the interfaces of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadInterfacesRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadInterfacesResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/interfaces", readInterfaces)

	// swagger:operation POST /read/count-interfaces read readCountInterfaces
	// ---
	// summary: Counts the interfaces of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadCountInterfacesRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadCountInterfacesResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/count-interfaces", readCountInterfaces)

	// swagger:operation POST /read/cpu-load read readCPULoad
	// ---
	// summary: Read out the CPU load of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadCPULoadRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadCPULoadResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/cpu-load", readCPULoad)

	// swagger:operation POST /read/memory-usage read readMemoryUsage
	// ---
	// summary: Read out the memory usage of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadMemoryUsageRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadMemoryUsageResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/memory-usage", readMemoryUsage)

	// swagger:operation POST /read/ups read readUPS
	// ---
	// summary: Reads out UPS data of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadUPSRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadUPSResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/ups", readUPS)

	// swagger:operation POST /read/sbc read readSBC
	// ---
	// summary: Reads out SBC data of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadSBCRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadSBCResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/sbc", readSBC)

	// swagger:operation POST /read/server read readServer
	// ---
	// summary: Reads out server data of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadServerRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadServerResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/server", readServer)

	// swagger:operation POST /read/disk read readDisk
	// ---
	// summary: Reads out disk data of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadDiskRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadDiskResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/disk", readDisk)

	// swagger:operation POST /read/hardware-health read readHardwareHealth
	// ---
	// summary: Reads out hardware health data of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadHardwareHealthRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadHardwareHealthResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/hardware-health", readHardwareHealth)

	// swagger:operation POST /read/high-availability read readHighAvailability
	// ---
	// summary: Read out the high availability status of a device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadHighAvailabilityRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadHighAvailabilityResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/high-availability", readHighAvailability)

	// swagger:operation POST /read/available-components read readAvailableComponents
	// ---
	// summary: Returns the available components for the device.
	// consumes:
	// - application/json
	// - application/xml
	// produces:
	// - application/json
	// - application/xml
	// parameters:
	// - name: body
	//   in: body
	//   description: Request to process.
	//   required: true
	//   schema:
	//     $ref: '#/definitions/ReadAvailableComponentsRequest'
	// responses:
	//   200:
	//     description: Returns the response.
	//     schema:
	//       $ref: '#/definitions/ReadAvailableComponentsResponse'
	//   400:
	//     description: Returns an error with more details in the body.
	//     schema:
	//       $ref: '#/definitions/OutputError'
	e.POST("/read/available-components", readAvailableComponents)

	// Start server
	go func() {
		var err error
		if viper.GetString("api.certfile") != "" && viper.GetString("api.keyfile") != "" {
			err = e.StartTLS(":"+viper.GetString("api.port"), viper.GetString("api.certfile"), viper.GetString("api.keyfile"))
		} else {
			err = e.Start(":" + viper.GetString("api.port"))
		}

		if dbErr := db.CloseConnection(ctx); dbErr != nil {
			log.Ctx(ctx).Err(dbErr).Msg("failed to close connection to the db")
		}

		if err != nil && err == http.ErrServerClosed {
			log.Ctx(ctx).Info().Msg("shutting down the server")
		} else {
			log.Ctx(ctx).Fatal().Err(err).Msg("unexpected server error")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Also close the connection to the database.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Ctx(ctx).Debug().Msg("received shutdown signal")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err = e.Shutdown(ctx); err != nil {
		log.Ctx(ctx).Fatal().Err(err).Msg("shutting down the server failed")
	}
}

func identify(ctx echo.Context) error {
	r := request.IdentifyRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkIdentify(ctx echo.Context) error {
	r := request.CheckIdentifyRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkSNMP(ctx echo.Context) error {
	r := request.CheckSNMPRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkInterfaceMetrics(ctx echo.Context) error {
	r := request.CheckInterfaceMetricsRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkTholaServer(ctx echo.Context) error {
	r := request.CheckTholaServerRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, nil)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkUPS(ctx echo.Context) error {
	r := request.CheckUPSRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkMemoryUsage(ctx echo.Context) error {
	r := request.CheckMemoryUsageRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkCPULoad(ctx echo.Context) error {
	r := request.CheckCPULoadRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkSBC(ctx echo.Context) error {
	r := request.CheckSBCRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkServer(ctx echo.Context) error {
	r := request.CheckServerRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkDisk(ctx echo.Context) error {
	r := request.CheckDiskRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkHardwareHealth(ctx echo.Context) error {
	r := request.CheckHardwareHealthRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func checkHighAvailability(ctx echo.Context) error {
	r := request.CheckHighAvailabilityRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readInterfaces(ctx echo.Context) error {
	r := request.ReadInterfacesRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readCountInterfaces(ctx echo.Context) error {
	r := request.ReadCountInterfacesRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readCPULoad(ctx echo.Context) error {
	r := request.ReadCPULoadRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readMemoryUsage(ctx echo.Context) error {
	r := request.ReadMemoryUsageRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readUPS(ctx echo.Context) error {
	r := request.ReadUPSRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readSBC(ctx echo.Context) error {
	r := request.ReadSBCRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readServer(ctx echo.Context) error {
	r := request.ReadServerRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readDisk(ctx echo.Context) error {
	r := request.ReadDiskRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readHardwareHealth(ctx echo.Context) error {
	r := request.ReadHardwareHealthRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readHighAvailability(ctx echo.Context) error {
	r := request.ReadHighAvailabilityRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func readAvailableComponents(ctx echo.Context) error {
	r := request.ReadAvailableComponentsRequest{}
	if err := ctx.Bind(&r); err != nil {
		return err
	}
	resp, err := handleAPIRequest(ctx, &r, &r.BaseRequest.DeviceData.IPAddress)
	if err != nil {
		return handleError(ctx, err)
	}
	return returnInFormat(ctx, http.StatusOK, resp)
}

func handleError(ctx echo.Context, err error) error {
	if tholaerr.IsNetworkError(err) {
		return returnInFormat(ctx, http.StatusBadRequest, tholaerr.OutputError{Error: "Network error: " + err.Error()})
	}
	if tholaerr.IsNotImplementedError(err) {
		return returnInFormat(ctx, http.StatusInternalServerError, tholaerr.OutputError{Error: "Function not implemented: " + err.Error()})
	}
	if tholaerr.IsNotFoundError(err) {
		return returnInFormat(ctx, http.StatusNotAcceptable, tholaerr.OutputError{Error: "Not found: " + err.Error()})
	}
	if tholaerr.IsTooManyRequestsError(err) {
		return returnInFormat(ctx, http.StatusTooManyRequests, tholaerr.OutputError{Error: "Too many requests: " + err.Error()})
	}
	return returnInFormat(ctx, http.StatusBadRequest, tholaerr.OutputError{Error: "Request failed: " + err.Error()})
}

func returnInFormat(ctx echo.Context, statusCode int, resp interface{}) error {
	if viper.GetString("api.format") == "json" {
		return ctx.JSON(statusCode, resp)
	} else if viper.GetString("api.format") == "xml" {
		return ctx.XML(statusCode, resp)
	}
	return ctx.String(http.StatusInternalServerError, "Invalid output format set")
}

func handleAPIRequest(echoCTX echo.Context, r request.Request, ip *string) (request.Response, error) {
	logger := log.With().Str("request_id", echoCTX.Request().Header.Get(echo.HeaderXRequestID)).Logger()
	ctx := logger.WithContext(context.Background())
	log.Ctx(ctx).Debug().Msg("incoming request")

	if ip != nil && !viper.GetBool("request.no-ip-lock") {
		ctx, cancel := request.CheckForTimeout(ctx, r)
		defer cancel()

		ch := getDeviceChannel(*ip)
		select {
		case <-ctx.Done():
			return r.HandlePreProcessError(errors.New("request timed out while waiting on the IP lock"))
		case <-ch:
			log.Ctx(ctx).Debug().Msgf("locked IP '%s'", *ip)
			defer func() {
				ch <- struct{}{}
				log.Ctx(ctx).Debug().Msgf("unlocked IP '%s'", *ip)
			}()
			return request.ProcessRequest(ctx, r)
		}
	} else {
		return request.ProcessRequest(ctx, r)
	}
}

func getDeviceChannel(ip string) chan struct{} {
	deviceChannels.RLock()
	ch, ok := deviceChannels.channels[ip]
	deviceChannels.RUnlock()
	if !ok {
		deviceChannels.Lock()
		if ch, ok = deviceChannels.channels[ip]; !ok {
			ch = make(chan struct{}, 1)
			ch <- struct{}{}
			deviceChannels.channels[ip] = ch
		}
		deviceChannels.Unlock()
	}
	return ch
}
