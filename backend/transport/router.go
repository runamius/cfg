package transport

import (
	"avito/iternal/repository/service"
	"avito/transport/handlers"
	"avito/transport/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	authSvc *service.AuthService,
	roomSvc *service.RoomService,
	scheduleSvc *service.ScheduleService,
	slotSvc *service.SlotService,
	bookingSvc *service.BookingService,
) *gin.Engine {
	r := gin.Default()

	infoH := &handlers.InfoHandler{}
	authH := handlers.NewAuthHandler(authSvc)
	roomH := handlers.NewRoomHandler(roomSvc)
	scheduleH := handlers.NewScheduleHandler(scheduleSvc)
	slotH := handlers.NewSlotHandler(slotSvc)
	bookingH := handlers.NewBookingHandler(bookingSvc)

	authMW := middleware.Auth(authSvc)
	adminMW := middleware.RequireRole("admin")
	userMW := middleware.RequireRole("user")

	r.GET("/_info", infoH.Info)
	r.POST("/dummyLogin", authH.DummyLogin)
	r.POST("/register", authH.Register)
	r.POST("/login", authH.Login)

	authed := r.Group("/", authMW)
	{
		authed.GET("/rooms/list", roomH.GetRooms)
		authed.GET("/rooms/:roomId/slots/list", slotH.ListSlots)
	}

	adminRoutes := r.Group("/", authMW, adminMW)
	{
		adminRoutes.POST("/rooms/create", roomH.CreateRoom)
		adminRoutes.POST("/rooms/:roomId/schedule/create", scheduleH.CreateSchedule)
		adminRoutes.GET("/bookings/list", bookingH.ListBookings)
	}

	userRoutes := r.Group("/", authMW, userMW)
	{
		userRoutes.POST("/bookings/create", bookingH.CreateBooking)
		userRoutes.GET("/bookings/my", bookingH.ListMyBookings)
		userRoutes.POST("/bookings/:bookingId/cancel", bookingH.CancelBooking)
	}

	return r
}
