// Package routes
package routes

import (
	"restaurant-backend/handlers"
	"restaurant-backend/middleware"
	"restaurant-backend/models"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	// Serve generated QR images
	// r.Static("/static", "./static")

	// r.GET("/", func(c *gin.Context) {
	// 	log.Printf("=== REQUEST ===")
	// 	log.Printf("Method: %s", c.Request.Method)
	// 	log.Printf("URL: %s", c.Request.URL.String())
	// 	log.Printf("Path: %s", c.Request.URL.Path)
	// 	log.Printf("RawPath: %s", c.Request.URL.RawPath)
	// 	log.Printf("RequestURI: %s", c.Request.RequestURI)
	// 	log.Printf("Host: %s", c.Request.Host)
	// })

	api := r.Group("/api/v1")

	// ---------- Public auth ----------
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// ---------- Public menu browsing (customer-facing, no auth) ----------
	public := api.Group("/public")
	{
		public.GET("/restaurants", handlers.ListRestaurants)
		public.GET("/restaurants/:id", handlers.GetRestaurant) // :id can be UUID or custom_sub_link
		public.GET("/restaurants/:id/categories", handlers.ListCategories)

		public.GET("/restaurants/:id/categories/:categoryId/foods", handlers.ListFoods)
		public.GET("/restaurants/:id/categories/:categoryId/foods/:foodId", handlers.GetFood)
		public.POST("/foods/:foodId/comments", handlers.AddComment)
	}

	// ---------- Authenticated user profile routes ----------
	userRoutes := api.Group("/user")
	userRoutes.Use(middleware.AuthRequired())
	{
		userRoutes.GET("/profile", handlers.GetProfile)
		userRoutes.PUT("/profile", handlers.UpdateProfile) // Update Name, Email, Password
	}

	// ---------- Authenticated restaurant owner management ----------
	restaurants := api.Group("/restaurants")
	restaurants.Use(middleware.AuthRequired())
	{
		// Restaurant CRUD
		restaurants.POST("", middleware.RequireRole(models.RoleRestaurantOwner, models.RoleSuperAdmin), handlers.CreateRestaurant)
		restaurants.GET("/:id", handlers.GetRestaurant)
		restaurants.PUT("/:id", handlers.UpdateRestaurant)
		restaurants.DELETE("/:id", handlers.DeleteRestaurant)

		// Category CRUD
		restaurants.GET("/:id/categories", handlers.ListCategories)
		restaurants.POST("/:id/categories", handlers.CreateCategory)
		restaurants.PUT("/:id/categories/:categoryId", handlers.UpdateCategory)
		restaurants.DELETE("/:id/categories/:categoryId", handlers.DeleteCategory)

		// Food CRUD
		restaurants.POST("/:id/categories/:categoryId/foods", handlers.CreateFood)
		restaurants.PUT("/:id/categories/:categoryId/foods/:foodId", handlers.UpdateFood)
		restaurants.DELETE("/:id/categories/:categoryId/foods/:foodId", handlers.DeleteFood)

		// Subscription Management
		restaurants.GET("/:id/subscription", handlers.GetSubscription)
		restaurants.POST("/:id/subscription/upgrade", handlers.UpgradeSubscription)

		// QR Code CRUD
		// restaurants.POST("/:id/qrcodes", handlers.CreateQR)
		// restaurants.GET("/:id/qrcodes", handlers.ListQRCodes)
		// restaurants.DELETE("/:id/qrcodes/:qrId", handlers.DeleteQR)
	}

	// ---------- Platform super_admin (Full Global Oversight & CRUD) ----------
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(), middleware.RequireRole(models.RoleSuperAdmin))
	{
		// Super Admin CRUD
		admin.PUT("/profile", handlers.UpdateSuperAdminProfile)

		// Users Global CRUD
		admin.GET("/users", handlers.ListUsers)
		admin.POST("/users", handlers.CreateUser)
		admin.PUT("/users/:id", handlers.UpdateUser)
		admin.PUT("/users/:id/role", handlers.UpdateUserRole)
		admin.DELETE("/users/:id", handlers.DeleteUser)

		// Restaurants Global Oversight (Reusing main handlers where appropriate)
		admin.GET("/restaurants", handlers.ListRestaurants)
		admin.DELETE("/restaurants/:id", handlers.DeleteRestaurant)

		// Subscriptions Global CRUD & Overrides
		admin.GET("/subscriptions", handlers.ListSubscriptions)
		admin.PUT("/restaurants/:id/subscription", handlers.UpdateSubscription)

		// Categories Global Views & Management
		admin.GET("/categories", handlers.ListCategories)
		admin.DELETE("/categories/:categoryId", handlers.DeleteCategory)

		// Foods Global Views & Management
		admin.GET("/foods", handlers.ListFoods)
		admin.DELETE("/foods/:foodId", handlers.DeleteFood)

		// Comments Moderation
		admin.GET("/comments", handlers.ListComments)
		admin.DELETE("/comments/:commentId", handlers.DeleteComment)

		// QR Codes Global Views & Management
		// admin.GET("/qrcodes", handlers.AdminListQRCodes)
		// admin.DELETE("/qrcodes/:qrId", handlers.DeleteQR)
	}
}
