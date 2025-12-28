package routes

import (
	"app/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	app.Get("/swagger/*", swagger.HandlerDefault)

	shipment_status_handlers := handlers.NewShipmentStatusHandler(db)
	shipment_method_handlers := handlers.NewShipmentMethodHandler(db)
	shipment_handler := handlers.NewShipmentHandler(db)
	request_status_handler := handlers.NewRequestStatusHandler(db)
	payment_status_handler := handlers.NewPaymentStatusHandler(db)
	payment_method_handler := handlers.NewPaymentMethodHandler(db)
	order_status_handler := handlers.NewOrderStatusHandler(db)
	payment_handler := handlers.NewPaymentHandler(db)
	user_handler := handlers.NewUserHandler(db)
	right_handler := handlers.NewRightHandler(db)
	order_handler := handlers.NewOrderHandler(db)
	order_item_handler := handlers.NewOrderItemHandler(db)
	request_handler := handlers.NewRequestHandler(db)
	category_handler := handlers.NewCategoryHandler(db)
	attribute_handler := handlers.NewAttributeHandler(db)
	manufacturer_handler := handlers.NewManufacturerHandler(db)
	product_handler := handlers.NewProductHandler(db)
	attribute_value_handler := handlers.NewAttributeValueHandler(db)
	product_image_handler := handlers.NewProductImageHandler(db)
	product_warehouse_handler := handlers.NewProductWarehouseHandler(db)
	common_block_handler := handlers.NewCommonBlockHandler(db)
	story_handler := handlers.NewStoryHandler(db)
	new_handler := handlers.NewNewHandler(db)

	shipmentStatusRouter := app.Group("/shipment-statuses")
	{
		shipmentStatusRouter.Get("/", shipment_status_handlers.GetAll)
		shipmentStatusRouter.Get("/:id", shipment_status_handlers.GetByID)
		shipmentStatusRouter.Post("/", shipment_status_handlers.Create)
		shipmentStatusRouter.Put("/:id", shipment_status_handlers.Update)
		shipmentStatusRouter.Patch("/:id", shipment_status_handlers.Patch)
		shipmentStatusRouter.Delete("/:id", shipment_status_handlers.Delete)
	}

	shipmentMethodRouter := app.Group("/shipment-methods")
	{
		shipmentMethodRouter.Get("/", shipment_method_handlers.GetAll)
		shipmentMethodRouter.Get("/:id", shipment_method_handlers.GetByID)
		shipmentMethodRouter.Post("/", shipment_method_handlers.Create)
		shipmentMethodRouter.Put("/:id", shipment_method_handlers.Update)
		shipmentMethodRouter.Patch("/:id", shipment_method_handlers.Patch)
		shipmentMethodRouter.Delete("/:id", shipment_method_handlers.Delete)
	}

	shipmentRouter := app.Group("/shipments")
	{
		shipmentRouter.Get("/", shipment_handler.GetAll)
		shipmentRouter.Get("/:id", shipment_handler.GetByID)
		shipmentRouter.Post("/", shipment_handler.Create)
		shipmentRouter.Put("/:id", shipment_handler.Update)
		shipmentRouter.Patch("/:id", shipment_handler.Patch)
		shipmentRouter.Delete("/:id", shipment_handler.Delete)
	}

	requestStatusRouter := app.Group("/request-statuses")
	{
		requestStatusRouter.Get("/", request_status_handler.GetAll)
		requestStatusRouter.Get("/:id", request_status_handler.GetByID)
		requestStatusRouter.Post("/", request_status_handler.Create)
		requestStatusRouter.Put("/:id", request_status_handler.Update)
		requestStatusRouter.Patch("/:id", request_status_handler.Patch)
		requestStatusRouter.Delete("/:id", request_status_handler.Delete)
	}

	paymentStatusRouter := app.Group("/payment-statuses")
	{
		paymentStatusRouter.Get("/", payment_status_handler.GetAll)
		paymentStatusRouter.Get("/:id", payment_status_handler.GetByID)
		paymentStatusRouter.Post("/", payment_status_handler.Create)
		paymentStatusRouter.Put("/:id", payment_status_handler.Update)
		paymentStatusRouter.Patch("/:id", payment_status_handler.Patch)
		paymentStatusRouter.Delete("/:id", payment_status_handler.Delete)
	}

	paymentMethodRouter := app.Group("/payment-methods")
	{
		paymentMethodRouter.Get("/", payment_method_handler.GetAll)
		paymentMethodRouter.Get("/:id", payment_method_handler.GetByID)
		paymentMethodRouter.Post("/", payment_method_handler.Create)
		paymentMethodRouter.Put("/:id", payment_method_handler.Update)
		paymentMethodRouter.Patch("/:id", payment_method_handler.Patch)
		paymentMethodRouter.Delete("/:id", payment_method_handler.Delete)
	}

	paymentRouter := app.Group("/payments")
	{
		paymentRouter.Get("/", payment_handler.GetAll)
		paymentRouter.Get("/:id", payment_handler.GetByID)
		paymentRouter.Post("/", payment_handler.Create)
		paymentRouter.Put("/:id", payment_handler.Update)
		paymentRouter.Patch("/:id", payment_handler.Patch)
		paymentRouter.Delete("/:id", payment_handler.Delete)
	}

	orderStatusRouter := app.Group("/order-statuses")
	{
		orderStatusRouter.Get("/", order_status_handler.GetAll)
		orderStatusRouter.Get("/:id", order_status_handler.GetByID)
		orderStatusRouter.Post("/", order_status_handler.Create)
		orderStatusRouter.Put("/:id", order_status_handler.Update)
		orderStatusRouter.Patch("/:id", order_status_handler.Patch)
		orderStatusRouter.Delete("/:id", order_status_handler.Delete)
	}

	rightRouter := app.Group("/rights")
	{
		rightRouter.Get("/", right_handler.GetAll)
		rightRouter.Get("/:id", right_handler.GetByID)
		rightRouter.Post("/", right_handler.Create)
		rightRouter.Put("/:id", right_handler.Update)
		rightRouter.Patch("/:id", right_handler.Patch)
		rightRouter.Delete("/:id", right_handler.Delete)
	}

	userRouter := app.Group("/users")
	{
		userRouter.Get("/", user_handler.GetAll)
		userRouter.Get("/:id", user_handler.GetByID)
		userRouter.Post("/", user_handler.Create)
		userRouter.Put("/:id", user_handler.Update)
		userRouter.Patch("/:id", user_handler.Patch)
		userRouter.Put("/:id/password", user_handler.UpdatePassword)
		userRouter.Delete("/:id", user_handler.Delete)
	}

	orderRouter := app.Group("/orders")
	{
		orderRouter.Get("/", order_handler.GetAll)
		orderRouter.Get("/:id", order_handler.GetByID)
		orderRouter.Post("/", order_handler.Create)
		orderRouter.Put("/:id", order_handler.Update)
		orderRouter.Patch("/:id", order_handler.Patch)
		orderRouter.Delete("/:id", order_handler.Delete)
	}

	orderItemRouter := app.Group("/order-items")
	{
		orderItemRouter.Get("/", order_item_handler.GetAll)
		orderItemRouter.Get("/:id", order_item_handler.GetByID)
		orderItemRouter.Post("/", order_item_handler.Create)
		orderItemRouter.Put("/:id", order_item_handler.Update)
		orderItemRouter.Patch("/:id", order_item_handler.Patch)
		orderItemRouter.Delete("/:id", order_item_handler.Delete)
	}

	requestRouter := app.Group("/requests")
	{
		requestRouter.Get("/", request_handler.GetAll)
		requestRouter.Get("/:id", request_handler.GetByID)
		requestRouter.Post("/", request_handler.Create)
		requestRouter.Put("/:id", request_handler.Update)
		requestRouter.Patch("/:id", request_handler.Patch)
		requestRouter.Delete("/:id", request_handler.Delete)
	}

	categoryRouter := app.Group("/categories")
	{
		categoryRouter.Get("/", category_handler.GetAll)
		categoryRouter.Get("/tree", category_handler.GetTree)
		categoryRouter.Get("/:id", category_handler.GetByID)
		categoryRouter.Get("/slug/:slug", category_handler.GetBySlug)
		categoryRouter.Post("/", category_handler.Create)
		categoryRouter.Put("/:id", category_handler.Update)
		categoryRouter.Patch("/:id", category_handler.Patch)
		categoryRouter.Delete("/:id", category_handler.Delete)
	}

	manufacturerRouter := app.Group("/manufacturers")
	{
		manufacturerRouter.Get("/", manufacturer_handler.GetAll)
		manufacturerRouter.Get("/:id", manufacturer_handler.GetByID)
		manufacturerRouter.Get("/slug/:slug", manufacturer_handler.GetBySlug)
		manufacturerRouter.Post("/", manufacturer_handler.Create)
		manufacturerRouter.Put("/:id", manufacturer_handler.Update)
		manufacturerRouter.Patch("/:id", manufacturer_handler.Patch)
		manufacturerRouter.Delete("/:id", manufacturer_handler.Delete)
	}

	attributeRouter := app.Group("/attributes")
	{
		attributeRouter.Get("/", attribute_handler.GetAll)
		attributeRouter.Get("/:id", attribute_handler.GetByID)
		attributeRouter.Get("/slug/:slug", attribute_handler.GetBySlug)
		attributeRouter.Post("/", attribute_handler.Create)
		attributeRouter.Put("/:id", attribute_handler.Update)
		attributeRouter.Patch("/:id", attribute_handler.Patch)
		attributeRouter.Delete("/:id", attribute_handler.Delete)
	}

	productRouter := app.Group("/products")
	{
		productRouter.Get("/", product_handler.GetAll)
		productRouter.Get("/:id", product_handler.GetByID)
		productRouter.Get("/slug/:slug", product_handler.GetBySlug)
		productRouter.Post("/", product_handler.Create)
		productRouter.Put("/:id", product_handler.Update)
		productRouter.Patch("/:id", product_handler.Patch)
		productRouter.Delete("/:id", product_handler.Delete)
	}

	attributeValueRouter := app.Group("/attribute-values")
	{
		attributeValueRouter.Get("/", attribute_value_handler.GetAll)
		attributeValueRouter.Get("/:id", attribute_value_handler.GetByID)
		attributeValueRouter.Post("/", attribute_value_handler.Create)
		attributeValueRouter.Put("/:id", attribute_value_handler.Update)
		attributeValueRouter.Patch("/:id", attribute_value_handler.Patch)
		attributeValueRouter.Delete("/:id", attribute_value_handler.Delete)
	}

	productImageRouter := app.Group("/product-images")
	{
		productImageRouter.Get("/", product_image_handler.GetAll)
		productImageRouter.Get("/:id", product_image_handler.GetByID)
		productImageRouter.Post("/", product_image_handler.Create)
		productImageRouter.Put("/:id", product_image_handler.Update)
		productImageRouter.Patch("/:id", product_image_handler.Patch)
		productImageRouter.Delete("/:id", product_image_handler.Delete)
	}

	productWarehouseRouter := app.Group("/product-warehouses")
	{
		productWarehouseRouter.Get("/", product_warehouse_handler.GetAll)
		productWarehouseRouter.Get("/:id", product_warehouse_handler.GetByID)
		productWarehouseRouter.Get("/product/:product_id", product_warehouse_handler.GetByProductID)
		productWarehouseRouter.Post("/", product_warehouse_handler.Create)
		productWarehouseRouter.Put("/:id", product_warehouse_handler.Update)
		productWarehouseRouter.Patch("/:id", product_warehouse_handler.Patch)
		productWarehouseRouter.Post("/:id/restock", product_warehouse_handler.Restock)
		productWarehouseRouter.Delete("/:id", product_warehouse_handler.Delete)
	}

	newRouter := app.Group("/news")
	{
		newRouter.Get("/", new_handler.GetAll)
		newRouter.Get("/:id", new_handler.GetByID)
		newRouter.Post("/", new_handler.Create)
		newRouter.Put("/:id", new_handler.Update)
		newRouter.Patch("/:id", new_handler.Patch)
		newRouter.Patch("/:id/toggle-visibility", new_handler.ToggleVisibility)
		newRouter.Delete("/:id", new_handler.Delete)
	}

	storyRouter := app.Group("/stories")
	{
		storyRouter.Get("/", story_handler.GetAll)
		storyRouter.Get("/:id", story_handler.GetByID)
		storyRouter.Post("/", story_handler.Create)
		storyRouter.Put("/:id", story_handler.Update)
		storyRouter.Patch("/:id", story_handler.Patch)
		storyRouter.Patch("/:id/toggle-visibility", story_handler.ToggleVisibility)
		storyRouter.Delete("/:id", story_handler.Delete)
	}

	commonBlockRouter := app.Group("/common-blocks")
	{
		commonBlockRouter.Get("/", common_block_handler.GetAll)
		commonBlockRouter.Get("/:id", common_block_handler.GetByID)
		commonBlockRouter.Post("/", common_block_handler.Create)
		commonBlockRouter.Put("/:id", common_block_handler.Update)
		commonBlockRouter.Patch("/:id", common_block_handler.Patch)
		commonBlockRouter.Patch("/:id/toggle-visibility", common_block_handler.ToggleVisibility)
		commonBlockRouter.Delete("/:id", common_block_handler.Delete)
	}
}
