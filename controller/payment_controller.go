package controller

import (
	"fmt"
	"shareway/helper"
	"shareway/infra/task"
	"shareway/infra/ws"
	"shareway/middleware"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentController struct {
	validate       *validator.Validate
	hub            *ws.Hub
	RideService    service.IRideService
	MapsService    service.IMapService
	UserService    service.IUsersService
	VehicleService service.IVehicleService
	PaymentService service.IPaymentService
	asyncClient    *task.AsyncClient
}

func NewPaymentController(validate *validator.Validate, hub *ws.Hub, rideService service.IRideService,
	mapService service.IMapService, userService service.IUsersService, vehicleService service.IVehicleService,
	paymentService service.IPaymentService,
	asyncClient *task.AsyncClient) *PaymentController {
	return &PaymentController{
		validate:       validate,
		hub:            hub,
		RideService:    rideService,
		MapsService:    mapService,
		UserService:    userService,
		VehicleService: vehicleService,
		PaymentService: paymentService,
		asyncClient:    asyncClient,
	}
}

// LinkMomoWallet godoc
// @Summary Link momo wallet to user account
// @Description Link momo wallet to user account
// @Tags payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body schemas.LinkMomoRequest true "Link momo wallet request"
// @Success 200 {object} schemas.LinkMomoWalletResponse "Link momo wallet response"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /payment/link-momo-wallet [post]
func (p *PaymentController) LinkMomoWallet(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))

	// Convert payload to map
	data, err := helper.ConvertToPayload(payload)

	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	var req schemas.LinkMomoRequest

	// Bind request to struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := p.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Get user info check if user already linked momo wallet
	user, err := p.UserService.GetUserByID(data.UserID)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get user info",
			"Không thể lấy thông tin người dùng",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	// If user already linked momo wallet, return error response
	if user.IsMomoLinked {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("user already linked momo wallet"),
			"User already linked momo wallet",
			"Người dùng đã liên kết ví momo",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Link momo wallet to user account
	momo, err := p.PaymentService.LinkMomoWallet(data.UserID, req.WalletPhoneNumber)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to link momo wallet",
			"Không thể liên kết ví momo",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	res := schemas.LinkMomoWalletResponse{
		Deeplink: momo.PayUrl, // Open browser to link momo wallet
	}

	response := helper.SuccessResponse(res, "Link momo wallet successfully", "Liên kết ví momo thành công")
	helper.GinResponse(ctx, 200, response)
}

// CheckoutRide when hitcher checkout ride with momo wallet and wait for status from IPN callback
// @Summary Checkout ride with momo
// CheckoutRide godoc
// @Description Checkout ride with momo
// @Tags payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body schemas.CheckoutRideRequest true "Checkout ride request"
// @Success 200 {object} helper.Response{data=object} "Link wallet response"
// @Failure 400 {object} helper.Response "Bad request"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /payment/checkout-ride [post]
func (p *PaymentController) CheckoutRide(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet((middleware.AuthorizationPayloadKey))
	// Convert payload to map
	data, err := helper.ConvertToPayload(payload)
	// If error occurs, return error response
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	var req schemas.CheckoutRideRequest
	// Bind request to struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to bind request",
			"Không thể bind request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Validate request
	if err := p.validate.Struct(req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to validate request",
			"Không thể validate request",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	// Perform checkout ride with momo wallet
	err = p.PaymentService.CheckoutRide(data.UserID, req)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to checkout ride",
			"Không thể thanh toán chuyến đi",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(nil, "Checkout ride successfully", "Thanh toán chuyến đi thành công")
	helper.GinResponse(ctx, 200, response)
}
